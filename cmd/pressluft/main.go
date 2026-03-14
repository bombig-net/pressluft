package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	command := "help"
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	var err error
	switch command {
	case "dev":
		err = runDev(args)
	case "build":
		err = runBuild(args)
	case "generate":
		err = runGenerate(args)
	case "status":
		err = runStatus(args)
	case "preflight":
		err = runPreflight(args)
	case "health":
		err = runHealth(args)
	case "stats":
		err = runStats(args)
	case "events":
		err = runEvents(args)
	case "reset":
		err = runReset(args)
	case "server-ssh":
		err = runServerSSH(args)
	case "test":
		err = runTest(args)
	case "format":
		err = runFormat()
	case "lint":
		err = runLint()
	case "validate":
		err = runValidate(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "pressluft: unknown command %q\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "pressluft %s: %v\n", command, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("pressluft — CLI for the Pressluft hosting panel")
	fmt.Println()
	fmt.Println("Development:")
	fmt.Println("  dev                      Start the local dev environment")
	fmt.Println("  build [server|agent]     Build binaries (default: full pipeline)")
	fmt.Println("  generate                 Regenerate TypeScript contracts from Go types")
	fmt.Println("  test [unit|integration]  Run test suites (default: all)")
	fmt.Println("  format                   Run go fmt on all packages")
	fmt.Println("  lint                     Run go vet on all packages")
	fmt.Println("  validate                 Run the full validation suite")
	fmt.Println()
	fmt.Println("Diagnostics:")
	fmt.Println("  status                   Inspect local dev state and callback durability")
	fmt.Println("  preflight                Validate local state before starting dev")
	fmt.Println("  health                   Verify runtime artifacts can be opened")
	fmt.Println("  stats                    Show row counts for key runtime tables")
	fmt.Println("  events [--limit N]       Show recent job events and activity rows")
	fmt.Println("  reset --force            Remove the local Pressluft state bundle")
	fmt.Println("  server-ssh TARGET        Print or execute SSH access for a managed server")
	fmt.Println()
	fmt.Println("Run 'pressluft <command> --help' for details on a specific command.")
}
