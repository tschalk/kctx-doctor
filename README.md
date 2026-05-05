# kctx-doctor

`kctx-doctor` is a small command-line tool for checking local Kubernetes
`kubeconfig` context wiring before a `kubectl` command needs it.

It reads local configuration only. It does not connect to the Kubernetes API
server, does not make changes to a cluster, and does not print credential
values.

## Install

```sh
go install github.com/tschalk/kctx-doctor/cmd/kctx-doctor@latest
```

## Usage

Check the current context from the default kubeconfig:

```sh
kctx-doctor
```

Check a specific file and context:

```sh
kctx-doctor --kubeconfig ./kubeconfig --context staging
```

Render machine-readable output:

```sh
kctx-doctor --output json
```

Use strict mode when warnings should fail a check:

```sh
kctx-doctor --strict
```

## Checks

The tool validates that:

- a kubeconfig file can be loaded
- the selected context exists
- the context references an existing cluster and user
- the cluster has a server configured
- outdated or unsafe configuration patterns are reported as warnings

The selected namespace is reported as informational context when it is not set.

## Exit Codes

- `0`: no failed checks
- `1`: failed checks, or warnings when `--strict` is enabled
- `2`: invalid command-line usage

