package main

import (
	"fmt"
	"os"
	"strings"

	"pressluft/internal/devdiag"
	"pressluft/internal/envconfig"
)

func runDoctor(args []string) error {
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			fmt.Println("pressluft doctor — check system health")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  --json  Output results as JSON")
			return nil
		case "--json":
			jsonOutput = true
		default:
			return fmt.Errorf("unknown flag %q", arg)
		}
	}

	runtime, err := resolveRuntime()
	if err != nil {
		return fmt.Errorf("resolve runtime: %w", err)
	}

	report := devdiag.Inspect(runtime)

	if jsonOutput {
		data, err := report.JSON()
		if err != nil {
			return fmt.Errorf("marshal report: %w", err)
		}
		os.Stdout.Write(data)
		fmt.Println()
	} else {
		printDoctorReport(report)
	}

	if !report.Healthy() {
		if !jsonOutput {
			fmt.Println()
			for _, issue := range report.Issues() {
				fmt.Printf("  - %s\n", issue)
			}
			fmt.Println()
			fmt.Println("To reset local state: rm -rf .pressluft")
		}
		os.Exit(1)
	}
	return nil
}

func resolveRuntime() (envconfig.ControlPlaneRuntime, error) {
	cwd, _ := os.Getwd()
	return envconfig.ResolveControlPlaneRuntime(true, cwd)
}

func printDoctorReport(report devdiag.Report) {
	fmt.Println("Pressluft doctor")
	fmt.Printf("  execution_mode: %s\n", report.Runtime.ExecutionMode)
	fmt.Printf("  data_dir: %s\n", report.Runtime.DataDir)
	fmt.Printf("  db_path: %s\n", report.Runtime.DBPath)
	fmt.Printf("  age_key_path: %s\n", report.Runtime.AgeKeyPath)
	fmt.Printf("  ca_key_path: %s\n", report.Runtime.CAKeyPath)
	fmt.Printf("  session_key_path: %s\n", report.Runtime.SessionSecretPath)
	if strings.TrimSpace(report.Runtime.ControlPlaneURL) == "" {
		fmt.Println("  callback_url: <unset>")
	} else {
		fmt.Printf("  callback_url: %s\n", report.Runtime.ControlPlaneURL)
	}
	fmt.Printf("  callback_url_mode: %s\n", report.CallbackURLMode)
	if report.DurableReconnectExpected {
		fmt.Println("  durable_reconnect: yes")
	} else {
		fmt.Println("  durable_reconnect: no")
	}
	fmt.Println()
	for _, check := range report.Checks {
		indicator := "?"
		switch check.Status {
		case devdiag.CheckStatusOK:
			indicator = "ok"
		case devdiag.CheckStatusWarning:
			indicator = "warn"
		case devdiag.CheckStatusError:
			indicator = "FAIL"
		}
		fmt.Printf("  [%s] %s: %s\n", indicator, check.Name, check.Detail)
	}
}
