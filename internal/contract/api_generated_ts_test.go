package contract

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGeneratedAPITypeScriptContractIsFresh(t *testing.T) {
	rendered, err := RenderAPITypeScriptModule()
	if err != nil {
		t.Fatalf("RenderAPITypeScriptModule() error = %v", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
	contractPath := filepath.Join(repoRoot, "web", "app", "lib", "api-contract.ts")
	current, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read generated api contract: %v", err)
	}

	if string(current) != rendered {
		t.Fatalf("generated api contract is stale; run `make generate-contract`")
	}
}
