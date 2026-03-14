package main

import (
	"fmt"
	"os"
	"path/filepath"

	"pressluft/internal/contract"
)

func runGenerate(args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			fmt.Println("pressluft generate — regenerate TypeScript contracts from Go types")
			fmt.Println()
			fmt.Println("Writes:")
			fmt.Println("  web/app/lib/platform-contract.generated.ts")
			fmt.Println("  web/app/lib/api-contract.ts")
			return nil
		}
	}

	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}

	// Platform contract.
	platformTS, err := contract.RenderTypeScriptModule()
	if err != nil {
		return fmt.Errorf("render platform contract: %w", err)
	}
	platformPath := filepath.Join(rootDir, "web", "app", "lib", "platform-contract.generated.ts")
	if err := os.WriteFile(platformPath, []byte(platformTS), 0o644); err != nil {
		return fmt.Errorf("write platform contract: %w", err)
	}

	// API contract.
	apiTS, err := contract.RenderAPITypeScriptModule()
	if err != nil {
		return fmt.Errorf("render api contract: %w", err)
	}
	apiPath := filepath.Join(rootDir, "web", "app", "lib", "api-contract.ts")
	if err := os.WriteFile(apiPath, []byte(apiTS), 0o644); err != nil {
		return fmt.Errorf("write api contract: %w", err)
	}

	fmt.Println("Contracts generated.")
	return nil
}
