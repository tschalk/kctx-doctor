package doctor

type Options struct {
	KubeconfigPath string
	Context        string
	Strict         bool
}

type Severity string

const (
	SeverityOK   Severity = "ok"
	SeverityInfo Severity = "info"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

type Report struct {
	Context string  `json:"context"`
	Status  Status  `json:"status"`
	Checks  []Check `json:"checks"`
}

func (r Report) ExitCode(strict bool) int {
	if r.Status == StatusFail {
		return 1
	}
	if strict && r.Status == StatusWarn {
		return 1
	}
	return 0
}
