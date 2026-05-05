package doctor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Diagnose(opts Options) Report {
	report := Report{Context: opts.Context}

	paths, explicitPath, err := loadingPaths(opts.KubeconfigPath)
	if err != nil {
		report.add("kubeconfig-path", SeverityFail, err.Error())
		report.finish()
		return report
	}

	if len(existingFiles(paths)) == 0 {
		report.add("kubeconfig-found", SeverityFail, "no kubeconfig file found")
		report.finish()
		return report
	}

	config, err := loadConfig(explicitPath)
	if err != nil {
		report.add("kubeconfig-load", SeverityFail, fmt.Sprintf("kubeconfig could not be loaded: %v", err))
		report.finish()
		return report
	}
	report.add("kubeconfig-load", SeverityOK, "kubeconfig loaded")

	selectedContext := strings.TrimSpace(opts.Context)
	if selectedContext == "" {
		selectedContext = strings.TrimSpace(config.CurrentContext)
	}
	report.Context = selectedContext

	if selectedContext == "" {
		report.add("context-selected", SeverityFail, "no context selected and kubeconfig has no current context")
		report.finish()
		return report
	}

	contextConfig, ok := config.Contexts[selectedContext]
	if !ok {
		report.add("context-exists", SeverityFail, fmt.Sprintf("context %q was not found", selectedContext))
		report.finish()
		return report
	}
	report.add("context-exists", SeverityOK, fmt.Sprintf("context %q exists", selectedContext))

	validateContext(&report, config, contextConfig)
	report.finish()
	return report
}

func validateContext(report *Report, config *clientcmdapi.Config, contextConfig *clientcmdapi.Context) {
	validateCluster(report, config, contextConfig.Cluster)
	validateUser(report, config, contextConfig.AuthInfo)

	if strings.TrimSpace(contextConfig.Namespace) == "" {
		report.add("namespace", SeverityInfo, "context does not set a namespace")
		return
	}
	report.add("namespace", SeverityOK, "context sets a namespace")
}

func validateCluster(report *Report, config *clientcmdapi.Config, clusterName string) {
	clusterName = strings.TrimSpace(clusterName)
	if clusterName == "" {
		report.add("cluster-reference", SeverityFail, "context does not reference a cluster")
		return
	}

	cluster, ok := config.Clusters[clusterName]
	if !ok {
		report.add("cluster-reference", SeverityFail, fmt.Sprintf("cluster %q was not found", clusterName))
		return
	}
	report.add("cluster-reference", SeverityOK, fmt.Sprintf("cluster %q exists", clusterName))

	server := strings.TrimSpace(cluster.Server)
	if server == "" {
		report.add("cluster-server", SeverityFail, "cluster server is not set")
		return
	}
	report.add("cluster-server", SeverityOK, "cluster server is set")

	if strings.HasPrefix(strings.ToLower(server), "http://") {
		report.add("cluster-server-scheme", SeverityWarn, "cluster server uses plain HTTP")
	}

	if cluster.InsecureSkipTLSVerify {
		report.add("cluster-tls", SeverityWarn, "cluster disables TLS certificate verification")
	}
}

func validateUser(report *Report, config *clientcmdapi.Config, userName string) {
	userName = strings.TrimSpace(userName)
	if userName == "" {
		report.add("user-reference", SeverityFail, "context does not reference a user")
		return
	}

	user, ok := config.AuthInfos[userName]
	if !ok {
		report.add("user-reference", SeverityFail, fmt.Sprintf("user %q was not found", userName))
		return
	}
	report.add("user-reference", SeverityOK, fmt.Sprintf("user %q exists", userName))

	if !hasAuthMaterial(user) {
		report.add("user-auth", SeverityFail, "user has no authentication configuration")
		return
	}
	report.add("user-auth", SeverityOK, "user has authentication configuration")

	if user.AuthProvider != nil {
		report.add("user-auth-provider", SeverityWarn, "user uses legacy auth-provider configuration")
	}

	if user.Exec != nil && strings.Contains(user.Exec.APIVersion, "v1alpha1") {
		report.add("user-exec-version", SeverityWarn, "user uses an old exec credential API version")
	}
}

func hasAuthMaterial(user *clientcmdapi.AuthInfo) bool {
	if user == nil {
		return false
	}

	return strings.TrimSpace(user.Token) != "" ||
		strings.TrimSpace(user.TokenFile) != "" ||
		strings.TrimSpace(user.Username) != "" ||
		strings.TrimSpace(user.Password) != "" ||
		strings.TrimSpace(user.ClientCertificate) != "" ||
		len(user.ClientCertificateData) > 0 ||
		strings.TrimSpace(user.ClientKey) != "" ||
		len(user.ClientKeyData) > 0 ||
		user.AuthProvider != nil ||
		user.Exec != nil
}

func loadConfig(explicitPath string) (*clientcmdapi.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if explicitPath != "" {
		rules.ExplicitPath = explicitPath
	}
	return rules.Load()
}

func loadingPaths(path string) ([]string, string, error) {
	if strings.TrimSpace(path) != "" {
		expanded, err := expandPath(path)
		if err != nil {
			return nil, "", err
		}
		return []string{expanded}, expanded, nil
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	paths := rules.GetLoadingPrecedence()
	if len(paths) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", errors.New("could not resolve home directory")
		}
		paths = []string{filepath.Join(home, ".kube", "config")}
	}

	return paths, "", nil
}

func existingFiles(paths []string) []string {
	existing := make([]string, 0, len(paths))
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			existing = append(existing, path)
		}
	}
	return existing
}

func expandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("kubeconfig path is empty")
	}

	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("could not resolve home directory")
		}
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("could not resolve home directory")
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}

	return path, nil
}

func (r *Report) add(id string, severity Severity, message string) {
	r.Checks = append(r.Checks, Check{
		ID:       id,
		Severity: severity,
		Message:  message,
	})
}

func (r *Report) finish() {
	status := StatusPass
	for _, check := range r.Checks {
		switch check.Severity {
		case SeverityFail:
			r.Status = StatusFail
			return
		case SeverityWarn:
			status = StatusWarn
		}
	}
	r.Status = status
}
