package platform

import (
	"fmt"
	"strings"
)

type ExecutionMode string

const (
	ExecutionModeDev                 ExecutionMode = "dev"
	ExecutionModeSingleNodeLocal     ExecutionMode = "single-node-local-control-plane"
	ExecutionModeProductionBootstrap ExecutionMode = "production-bootstrap"
	defaultControlPlaneExecutionMode ExecutionMode = ExecutionModeSingleNodeLocal
	defaultNonDevAgentExecutionMode  ExecutionMode = ExecutionModeProductionBootstrap
)

type SupportLevel string

const (
	SupportLevelSupported    SupportLevel = "supported"
	SupportLevelExperimental SupportLevel = "experimental"
	SupportLevelUnavailable  SupportLevel = "unavailable"
)

type ServerStatus string

type SetupState string

const (
	ServerStatusPending      ServerStatus = "pending"
	ServerStatusProvisioning ServerStatus = "provisioning"
	ServerStatusConfiguring  ServerStatus = "configuring"
	ServerStatusRebuilding   ServerStatus = "rebuilding"
	ServerStatusResizing     ServerStatus = "resizing"
	ServerStatusDeleting     ServerStatus = "deleting"
	ServerStatusDeleted      ServerStatus = "deleted"
	ServerStatusReady        ServerStatus = "ready"
	ServerStatusFailed       ServerStatus = "failed"
)

const (
	SetupStateNotStarted SetupState = "not_started"
	SetupStateRunning    SetupState = "running"
	SetupStateDegraded   SetupState = "degraded"
	SetupStateReady      SetupState = "ready"
)

type CallbackURLMode string

const (
	CallbackURLModeUnknown   CallbackURLMode = "unknown"
	CallbackURLModeStable    CallbackURLMode = "stable"
	CallbackURLModeEphemeral CallbackURLMode = "ephemeral"
)

func QueuedServerStatusForJobKind(kind string) (ServerStatus, bool) {
	switch strings.TrimSpace(kind) {
	case "delete_server":
		return ServerStatusDeleting, true
	case "rebuild_server":
		return ServerStatusRebuilding, true
	case "resize_server":
		return ServerStatusResizing, true
	default:
		return "", false
	}
}

func IsDeletingOrDeletedServerStatus(status string) bool {
	switch ServerStatus(strings.TrimSpace(status)) {
	case ServerStatusDeleting, ServerStatusDeleted:
		return true
	default:
		return false
	}
}

func DetectCallbackURLMode(raw string) CallbackURLMode {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return CallbackURLModeUnknown
	}
	if strings.Contains(value, ".trycloudflare.com") {
		return CallbackURLModeEphemeral
	}
	return CallbackURLModeStable
}

func NormalizeControlPlaneExecutionMode(raw string, devBuild bool) (ExecutionMode, error) {
	if devBuild {
		if strings.TrimSpace(raw) != "" && ExecutionMode(strings.TrimSpace(raw)) != ExecutionModeDev {
			return "", fmt.Errorf("dev builds only support execution mode %q", ExecutionModeDev)
		}
		return ExecutionModeDev, nil
	}

	mode := ExecutionMode(strings.TrimSpace(raw))
	if mode == "" {
		mode = defaultControlPlaneExecutionMode
	}

	switch mode {
	case ExecutionModeSingleNodeLocal, ExecutionModeProductionBootstrap:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported control-plane execution mode %q", mode)
	}
}

func NormalizeAgentExecutionMode(raw string, devBuild bool) (ExecutionMode, error) {
	if devBuild {
		if strings.TrimSpace(raw) != "" && ExecutionMode(strings.TrimSpace(raw)) != ExecutionModeDev {
			return "", fmt.Errorf("dev agent builds only support execution mode %q", ExecutionModeDev)
		}
		return ExecutionModeDev, nil
	}

	mode := ExecutionMode(strings.TrimSpace(raw))
	if mode == "" {
		mode = defaultNonDevAgentExecutionMode
	}

	switch mode {
	case ExecutionModeProductionBootstrap:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported agent execution mode %q", mode)
	}
}
