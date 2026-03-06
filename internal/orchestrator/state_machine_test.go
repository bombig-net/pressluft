package orchestrator

import "testing"

func TestCanTransition(t *testing.T) {
	if !CanTransition(JobStatusQueued, JobStatusRunning) {
		t.Fatal("expected queued -> running to be valid")
	}
	if CanTransition(JobStatusSucceeded, JobStatusRunning) {
		t.Fatal("expected succeeded -> running to be invalid")
	}
}

func TestValidateTransition(t *testing.T) {
	if err := ValidateTransition(JobStatusRunning, JobStatusSucceeded); err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}
	if err := ValidateTransition(JobStatusRunning, JobStatusQueued); err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestIsTerminalStatus(t *testing.T) {
	if !IsTerminalStatus(JobStatusSucceeded) {
		t.Fatal("expected succeeded to be terminal")
	}
	if IsTerminalStatus(JobStatusRunning) {
		t.Fatal("expected running to be non-terminal")
	}
}

func TestAllowedStatusesForKind(t *testing.T) {
	statuses := AllowedStatusesForKind(string(JobKindDeleteServer))
	if len(statuses) != 4 {
		t.Fatalf("len(statuses) = %d, want 4", len(statuses))
	}
	if statuses[0] != JobStatusQueued || statuses[len(statuses)-1] != JobStatusFailed {
		t.Fatalf("unexpected statuses: %#v", statuses)
	}
}

func TestJobKindPolicyIncludesTimeoutAndRecovery(t *testing.T) {
	policy, ok := JobKindPolicy(string(JobKindRestartService))
	if !ok {
		t.Fatal("expected restart_service policy")
	}
	if policy.Timeout <= 0 {
		t.Fatal("expected positive timeout")
	}
	if policy.RetryLimit != 0 {
		t.Fatalf("retry_limit = %d, want 0", policy.RetryLimit)
	}
	if policy.Recovery == "" {
		t.Fatal("expected recovery guidance")
	}
}

func TestIsKnownJobKind(t *testing.T) {
	if !IsKnownJobKind(string(JobKindProvisionServer)) {
		t.Fatal("expected provision_server to be known")
	}
	if IsKnownJobKind("unknown") {
		t.Fatal("expected unknown kind to be rejected")
	}
}
