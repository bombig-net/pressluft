package contract

import (
	"strings"
	"testing"
)

func TestRenderTypeScriptModuleExportsContractSurface(t *testing.T) {
	rendered, err := RenderTypeScriptModule()
	if err != nil {
		t.Fatalf("RenderTypeScriptModule() error = %v", err)
	}
	for _, needle := range []string{
		"export const platformContract = ",
		"export type ServerStatus = ",
		"export const jobKindLabels: Record<JobKind, string>",
		"export const jobTerminalStatuses: readonly JobTerminalStatus[] = ",
	} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("RenderTypeScriptModule() missing %q", needle)
		}
	}
}

func TestJobKindsRemainSorted(t *testing.T) {
	spec := SpecData()
	for i := 1; i < len(spec.JobKinds); i++ {
		if spec.JobKinds[i-1].Kind > spec.JobKinds[i].Kind {
			t.Fatalf("job kinds are not sorted: %q appears before %q", spec.JobKinds[i-1].Kind, spec.JobKinds[i].Kind)
		}
	}
}
