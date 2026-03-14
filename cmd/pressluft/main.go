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
	case "test":
		err = runTest(args)
	case "doctor":
		err = runDoctor(args)
	case "server-ssh":
		err = runServerSSH(args)
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
	fmt.Println("Commands:")
	fmt.Println("  dev                      Start the local dev environment")
	fmt.Println("  build [server|agent]     Build binaries (default: full pipeline)")
	fmt.Println("  test [unit|integration]  Run test suites (default: all)")
	fmt.Println("  doctor [--json]          Check system health")
	fmt.Println("  server-ssh TARGET        SSH access for a managed server")
	fmt.Println()
	fmt.Println("Run 'pressluft <command> --help' for details on a specific command.")
}
