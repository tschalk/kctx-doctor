package doctor

import (
	"fmt"
	"io"
	"strings"
)

func WriteText(w io.Writer, report Report) {
	context := report.Context
	if strings.TrimSpace(context) == "" {
		context = "<none>"
	}

	fmt.Fprintf(w, "kctx-doctor: %s for context %q\n", strings.ToUpper(string(report.Status)), context)
	for _, check := range report.Checks {
		fmt.Fprintf(w, "[%s] %s\n", strings.ToUpper(string(check.Severity)), check.Message)
	}
}
