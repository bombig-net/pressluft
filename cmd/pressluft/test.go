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
