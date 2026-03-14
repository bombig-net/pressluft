package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var unitTestPackages = []string{
	"./cmd/pressluft-server",
	"./cmd/pressluft-agent",
	"./internal/agent",
	"./internal/agentauth",
	"./internal/agentcommand",
	"./internal/auth",
	"./internal/contract",
	"./internal/database",
	"./internal/dispatch",
	"./internal/envconfig",
	"./internal/platform",
	"./internal/registration",
	"./internal/security",
	"./internal/worker",
	"./internal/ws",
}

var integrationTestPackages = []string{
	"./internal/activity",
	"./internal/orchestrator",
	"./internal/provider/...",
	"./internal/runner/ansible",
	"./internal/server",
	"./internal/server/profiles",
}

func runTest(args []string) error {
	suite := ""
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			fmt.Println("pressluft test [unit|integration] — run test suites")
			fmt.Println()
			fmt.Println("  pressluft test              Run all tests (unit + integration)")
			fmt.Println("  pressluft test unit         Run unit tests only")
			fmt.Println("  pressluft test integration  Run integration tests only")
			return nil
		case "unit", "integration":
			suite = arg
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag %q", arg)
			}
			return fmt.Errorf("unknown test suite %q (use unit or integration)", arg)
		}
	}

	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}

	switch suite {
	case "unit":
		return goTest(rootDir, unitTestPackages, false)
	case "integration":
		return goTest(rootDir, integrationTestPackages, true)
	default:
		if err := goTest(rootDir, unitTestPackages, false); err != nil {
			return err
		}
		return goTest(rootDir, integrationTestPackages, true)
	}
}

func runFormat() error {
	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command(goCmd(), "fmt", "./...")
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runLint() error {
	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command(goCmd(), "vet", "./...")
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runValidate(args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			fmt.Println("pressluft validate — run the full validation suite")
			fmt.Println()
			fmt.Println("Runs: format check, lint, all tests, ansible syntax, profile validation, and full build")
			return nil
		}
	}

	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}

	// Format check.
	fmt.Println("Checking formatting...")
	fmtCheck := exec.Command("gofmt", "-l", "cmd", "internal")
	fmtCheck.Dir = rootDir
	fmtOut, err := fmtCheck.Output()
	if err != nil {
		return fmt.Errorf("gofmt check failed: %w", err)
	}
	if len(strings.TrimSpace(string(fmtOut))) > 0 {
		fmt.Println(string(fmtOut))
		return fmt.Errorf("unformatted files found")
	}

	// Lint.
	fmt.Println("Running lint...")
	if err := runLint(); err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	// Tests.
	fmt.Println("Running tests...")
	if err := runTest(nil); err != nil {
		return fmt.Errorf("test: %w", err)
	}

	// Ansible syntax.
	fmt.Println("Checking ansible syntax...")
	if err := ansibleSyntaxCheck(rootDir); err != nil {
		return fmt.Errorf("ansible syntax: %w", err)
	}

	// Profile validation.
	fmt.Println("Validating profiles...")
	if err := goTest(rootDir, []string{"./internal/server/profiles"}, true); err != nil {
		return fmt.Errorf("profile validation: %w", err)
	}

	// Full build.
	fmt.Println("Running full build...")
	if err := buildFull(rootDir); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	fmt.Println("Validation complete.")
	return nil
}

func goTest(rootDir string, packages []string, noCache bool) error {
	testArgs := []string{"test"}
	if noCache {
		testArgs = append(testArgs, "-count=1")
	}
	testArgs = append(testArgs, packages...)

	cmd := exec.Command(goCmd(), testArgs...)
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var ansiblePlaybooks = []string{
	"ops/ansible/playbooks/deploy-site.yml",
	"ops/ansible/playbooks/hetzner/delete.yml",
	"ops/ansible/playbooks/hetzner/firewalls.yml",
	"ops/ansible/playbooks/hetzner/provision.yml",
	"ops/ansible/playbooks/hetzner/rebuild.yml",
	"ops/ansible/playbooks/hetzner/resize.yml",
	"ops/ansible/playbooks/hetzner/volume.yml",
}

func ansibleSyntaxCheck(rootDir string) error {
	for _, playbook := range ansiblePlaybooks {
		cmd := exec.Command("ansible-playbook", "-i", "localhost,", "-c", "local", "--syntax-check", playbook)
		cmd.Dir = rootDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("syntax check %s: %w", playbook, err)
		}
	}
	return nil
}
