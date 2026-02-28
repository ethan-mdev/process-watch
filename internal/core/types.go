package core

// Process represents a system process.
type Process struct {
	Name          string  `json:"name"`
	PID           int32   `json:"pid"`
	State         string  `json:"state"` // e.g., "running", "stopped"
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryMB      float64 `json:"memoryMB"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
}

// WatchlistItem represents a process being monitored.
type WatchlistItem struct {
	Name         string `json:"name"`
	RestartCmd   string `json:"restartCmd"`
	AutoRestart  bool   `json:"autoRestart"`
	MaxRetries   int    `json:"maxRetries"`
	CooldownSecs int    `json:"cooldownSecs"`
	RestartCount int    `json:"restartCount"`
	FailCount    int    `json:"failCount"`
	LastRestart  string `json:"lastRestart"`
}

// WatchStatus (central data type flowing from watcher -> TUI and watcher -> prometheus)
type WatchStatus struct {
	Entry             WatchlistItem `json:"entry"`
	Process           *Process      `json:"process,omitempty"` // nil if not running
	Running           bool          `json:"running"`
	InCooldown        bool          `json:"inCooldown"`
	CooldownRemaining int           `json:"cooldownRemaining"` // seconds
}

// Event represents a loggable event in the system.
type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
