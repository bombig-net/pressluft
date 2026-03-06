package orchestrator

import "time"

// JobStatus is the durable lifecycle state for orchestration jobs.
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

const (
	JobEventTypeCreated      = "job_created"
	JobEventTypeStepStarted  = "step_started"
	JobEventTypeStepComplete = "step_completed"
	JobEventTypeCommandLog   = "command_log"
	JobEventTypeSucceeded    = "job_succeeded"
	JobEventTypeFailed       = "job_failed"
	JobEventTypeRecovered    = "job_recovered"
	JobEventTypeTimedOut     = "job_timed_out"
)

// JobKind is the canonical identifier for a supported orchestration workflow.
type JobKind string

const (
	JobKindProvisionServer JobKind = "provision_server"
	JobKindDeleteServer    JobKind = "delete_server"
	JobKindRebuildServer   JobKind = "rebuild_server"
	JobKindResizeServer    JobKind = "resize_server"
	JobKindUpdateFirewalls JobKind = "update_firewalls"
	JobKindManageVolume    JobKind = "manage_volume"
	JobKindRestartService  JobKind = "restart_service"
)

type JobKindSpec struct {
	Kind            JobKind
	Label           string
	AllowedStatuses []JobStatus
	Destructive     bool
	Experimental    bool
	Timeout         time.Duration
	RetryLimit      int
	Recovery        string
}

var supportedJobKinds = []JobKindSpec{
	{Kind: JobKindProvisionServer, Label: "Server provisioning", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Timeout: 45 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; inspect provider state before retrying manually"},
	{Kind: JobKindDeleteServer, Label: "Server deletion", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Destructive: true, Experimental: true, Timeout: 20 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; verify provider-side deletion before retrying manually"},
	{Kind: JobKindRebuildServer, Label: "Server rebuild", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Destructive: true, Experimental: true, Timeout: 45 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; inspect machine state before retrying manually"},
	{Kind: JobKindResizeServer, Label: "Server resize", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Destructive: true, Experimental: true, Timeout: 20 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; inspect provider-side resize state before retrying manually"},
	{Kind: JobKindUpdateFirewalls, Label: "Firewall update", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Experimental: true, Timeout: 15 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; retry manually after inspection"},
	{Kind: JobKindManageVolume, Label: "Volume management", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Experimental: true, Timeout: 20 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption; retry manually after inspection"},
	{Kind: JobKindRestartService, Label: "Service restart", AllowedStatuses: []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed}, Experimental: true, Timeout: 2 * time.Minute, RetryLimit: 0, Recovery: "mark failed on worker interruption or timeout; late agent results are ignored"},
}

// SupportedJobKinds returns the current canonical job-kind contract.
func SupportedJobKinds() []JobKindSpec {
	out := make([]JobKindSpec, len(supportedJobKinds))
	copy(out, supportedJobKinds)
	return out
}

// IsKnownJobKind reports whether kind is part of the current runtime contract.
func IsKnownJobKind(kind string) bool {
	_, ok := JobKindPolicy(kind)
	return ok
}

// JobKindLabel returns a human-readable label for a supported job kind.
func JobKindLabel(kind string) string {
	spec, ok := JobKindPolicy(kind)
	if !ok {
		return kind
	}
	return spec.Label
}

// AllowedStatusesForKind returns the lifecycle states currently used by the runtime for kind.
func AllowedStatusesForKind(kind string) []JobStatus {
	spec, ok := JobKindPolicy(kind)
	if !ok {
		return nil
	}
	out := make([]JobStatus, len(spec.AllowedStatuses))
	copy(out, spec.AllowedStatuses)
	return out
}

func JobKindPolicy(kind string) (JobKindSpec, bool) {
	for _, spec := range supportedJobKinds {
		if string(spec.Kind) == kind {
			return spec, true
		}
	}
	return JobKindSpec{}, false
}

// Job is the persisted orchestration unit.
type Job struct {
	ID          int64     `json:"id"`
	ServerID    int64     `json:"server_id,omitempty"`
	Kind        string    `json:"kind"`
	Status      JobStatus `json:"status"`
	CurrentStep string    `json:"current_step"`
	RetryCount  int       `json:"retry_count"`
	LastError   string    `json:"last_error,omitempty"`
	Payload     string    `json:"payload,omitempty"`
	StartedAt   string    `json:"started_at,omitempty"`
	FinishedAt  string    `json:"finished_at,omitempty"`
	TimeoutAt   string    `json:"timeout_at,omitempty"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	CommandID   *string   `json:"command_id,omitempty"`
}

// JobEvent is an ordered event entry consumed by the dashboard.
type JobEvent struct {
	JobID      int64  `json:"job_id"`
	Seq        int64  `json:"seq"`
	EventType  string `json:"event_type"`
	Level      string `json:"level"`
	StepKey    string `json:"step_key,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message"`
	Payload    string `json:"payload,omitempty"`
	OccurredAt string `json:"occurred_at"`
}

// CreateJobInput is the job creation payload.
type CreateJobInput struct {
	Kind     string
	ServerID int64
	Payload  string
}

// TransitionInput updates a job lifecycle state.
type TransitionInput struct {
	ToStatus    JobStatus
	CurrentStep string
	LastError   string
	RetryCount  int
}

// CreateEventInput appends an event for a job timeline.
type CreateEventInput struct {
	EventType string
	Level     string
	StepKey   string
	Status    string
	Message   string
	Payload   string
}
