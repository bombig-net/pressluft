package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runBuild(args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			fmt.Println("pressluft build — build project binaries")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  pressluft build              Full pipeline: generate contracts, build frontend, embed, compile server + agent")
			fmt.Println("  pressluft build server       Build only the control-plane server binary")
			fmt.Println("  pressluft build agent        Build the production agent binary")
			fmt.Println("  pressluft build agent --dev  Build the dev agent binary")
			return nil
		}
	}

	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}

	target := ""
	devMode := false
	for _, arg := range args {
		switch arg {
		case "--dev":
			devMode = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag %q", arg)
			}
			if target != "" {
				return fmt.Errorf("build accepts at most one target (server or agent)")
			}
			target = arg
		}
	}

	switch target {
	case "":
		return buildFull(rootDir)
	case "server":
		return buildServer(rootDir)
	case "agent":
		return buildAgent(rootDir, devMode)
	default:
		return fmt.Errorf("unknown build target %q (use server or agent)", target)
	}
}

func buildFull(rootDir string) error {
	fmt.Println("Generating contracts...")
	if err := runGenerate(nil); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	fmt.Println("Building frontend...")
	if err := buildFrontend(rootDir); err != nil {
		return fmt.Errorf("frontend: %w", err)
	}

	fmt.Println("Embedding frontend assets...")
	if err := embedFrontend(rootDir); err != nil {
		return fmt.Errorf("embed: %w", err)
	}

	fmt.Println("Building server...")
	if err := buildServer(rootDir); err != nil {
		return fmt.Errorf("server: %w", err)
	}

	fmt.Println("Building agent...")
	if err := buildAgent(rootDir, false); err != nil {
		return fmt.Errorf("agent: %w", err)
	}

	fmt.Println("Build complete.")
	return nil
}

func buildServer(rootDir string) error {
	binPath := filepath.Join(rootDir, "bin", "pressluft-server")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command(goCmd(), "build", "-o", binPath, "./cmd/pressluft-server")
	cmd.Dir = rootDir
	cmd.Env = appendBuildEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildAgent(rootDir string, dev bool) error {
	binPath := filepath.Join(rootDir, "bin", "pressluft-agent")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		return err
	}
	buildArgs := []string{"build", "-o", binPath}
	if dev {
		buildArgs = append(buildArgs, "-tags", "dev")
	}
	buildArgs = append(buildArgs, "./cmd/pressluft-agent")
	cmd := exec.Command(goCmd(), buildArgs...)
	cmd.Dir = rootDir
	cmd.Env = appendBuildEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildFrontend(rootDir string) error {
	webDir := filepath.Join(rootDir, "web")

	// Install deps if needed.
	if _, err := os.Stat(filepath.Join(webDir, "node_modules")); os.IsNotExist(err) {
		install := exec.Command(npmCmd(), "--prefix", webDir, "install")
		install.Dir = rootDir
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if err := install.Run(); err != nil {
			return fmt.Errorf("install frontend deps: %w", err)
		}
	}

	gen := exec.Command(npmCmd(), "--prefix", webDir, "run", "generate")
	gen.Dir = rootDir
	gen.Env = append(os.Environ(), "NODE_OPTIONS=--max-old-space-size=8192")
	gen.Stdout = os.Stdout
	gen.Stderr = os.Stderr
	if err := gen.Run(); err != nil {
		return fmt.Errorf("generate frontend: %w", err)
	}

	// Verify output exists.
	indexPath := filepath.Join(webDir, ".output", "public", "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("frontend build did not produce %s", indexPath)
	}
	return nil
}

func embedFrontend(rootDir string) error {
	embedDir := filepath.Join(rootDir, "internal", "server", "dist")
	if err := os.RemoveAll(embedDir); err != nil {
		return err
	}
	if err := os.MkdirAll(embedDir, 0o755); err != nil {
		return err
	}

	// Create .gitkeep.
	if err := os.WriteFile(filepath.Join(embedDir, ".gitkeep"), nil, 0o644); err != nil {
		return err
	}

	// Copy built frontend assets.
	srcDir := filepath.Join(rootDir, "web", ".output", "public")
	cmd := exec.Command("cp", "-R", srcDir+"/.", embedDir+"/")
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goCmd() string {
	if v := os.Getenv("GO"); v != "" {
		return v
	}
	return "go"
}

func npmCmd() string {
	if v := os.Getenv("NPM"); v != "" {
		return v
	}
	return "pnpm"
}
