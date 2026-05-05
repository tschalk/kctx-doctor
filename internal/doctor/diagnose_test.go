package doctor

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestDiagnoseValidConfigPasses(t *testing.T) {
	report := Diagnose(Options{
		KubeconfigPath: filepath.Join("..", "..", "testdata", "valid.yaml"),
	})

	if report.Status != StatusPass {
		t.Fatalf("status = %s, want %s: %#v", report.Status, StatusPass, report.Checks)
	}
	if report.Context != "dev" {
		t.Fatalf("context = %q, want dev", report.Context)
	}
	if code := report.ExitCode(false); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestDiagnoseMissingContextFails(t *testing.T) {
	report := Diagnose(Options{
		KubeconfigPath: filepath.Join("..", "..", "testdata", "valid.yaml"),
		Context:        "missing",
	})

	if report.Status != StatusFail {
		t.Fatalf("status = %s, want %s", report.Status, StatusFail)
	}
	if code := report.ExitCode(false); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestDiagnoseWarnConfigHonorsStrictMode(t *testing.T) {
	report := Diagnose(Options{
		KubeconfigPath: filepath.Join("..", "..", "testdata", "warn.yaml"),
	})

	if report.Status != StatusWarn {
		t.Fatalf("status = %s, want %s: %#v", report.Status, StatusWarn, report.Checks)
	}
	if code := report.ExitCode(false); code != 0 {
		t.Fatalf("exit code without strict = %d, want 0", code)
	}
	if code := report.ExitCode(true); code != 1 {
		t.Fatalf("exit code with strict = %d, want 1", code)
	}
}

func TestDiagnoseMissingClusterFails(t *testing.T) {
	report := Diagnose(Options{
		KubeconfigPath: filepath.Join("..", "..", "testdata", "missing-cluster.yaml"),
	})

	if report.Status != StatusFail {
		t.Fatalf("status = %s, want %s", report.Status, StatusFail)
	}
}

func TestReportJSONShape(t *testing.T) {
	report := Diagnose(Options{
		KubeconfigPath: filepath.Join("..", "..", "testdata", "valid.yaml"),
	})

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(report); err != nil {
		t.Fatalf("encode report: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if decoded.Context != report.Context || decoded.Status != report.Status || len(decoded.Checks) == 0 {
		t.Fatalf("decoded report mismatch: %#v", decoded)
	}
}
