# Runtime Drivers

OpenBerth's deploy backend is pluggable. The built-in **docker** driver shells
out to the `docker` CLI, but contributors can add drivers for Kubernetes,
Firecracker microVMs, or any other orchestration layer that satisfies the
runtime contract below.

Drivers are **in-tree and compile-time**: adding one means dropping a package
into `apps/server/internal/runtime/<name>/` that implements the `Runtime`
interface and calls `runtime.Register` in its `init()`. The server binary
includes every driver that `apps/server/main.go` blank-imports. Driver
selection at runtime happens via `cfg.Runtime.Driver` in `config.json`
(default: `"docker"`).

## The Interface

All drivers implement the twelve methods on
`apps/server/internal/runtime.Runtime`:

```go
type Runtime interface {
    Deploy(DeployOpts) (*Result, error)          // fresh build + run
    Rebuild(DeployOpts) (*Result, error)         // blue-green update
    RestartRuntime(DeployOpts) (*Result, error)  // re-launch runtime only (secret rotation)
    StartSandbox(SandboxOpts) (*Result, error)   // long-lived dev instance

    Destroy(instanceID string) error
    Status(instanceID string) Status
    Restart(instanceID string) bool
    Port(instanceID string) int

    Exec(instanceID, command string, timeout time.Duration) (ExecResult, error)
    Logs(instanceID string, tail int) string
    LogStream(instanceID string, tail int) (io.ReadCloser, error)

    Capabilities() Capabilities
}
```

Return types (`Result`, `Endpoint`, `Status`, `ExecResult`, `Capabilities`)
carry no Docker concepts — see `runtime/runtime.go` for their full
definitions.

## The Environment Contract

Regardless of backend, framework providers (`internal/framework/*`) emit
POSIX shell scripts for build and run. Every driver must give those scripts
the same environment:

| Path / Variable | Mode      | Purpose                                    |
|-----------------|-----------|--------------------------------------------|
| `/bin/sh`       | present   | Build / run scripts begin with `#!/bin/sh` |
| `/app`          | rw        | Workdir. Build artifacts live here.        |
| `/app/code`     | ro        | Source tree bind-mount.                    |
| `/data`         | rw        | Persistent per-instance state.             |
| `/tmp`          | rw tmpfs  | Scratch. Size-limited by the driver.       |
| `$PORT`         | injected  | Port the app must listen on.               |
| `$DATA_DIR`     | injected  | Set to `/data`.                            |

For sandboxes (dev mode), `/app` is mounted **rw** from the host so that
`berth_sandbox_push` writes land immediately. For deploys it's a named
volume (Docker driver) or equivalent artifact store.

Framework-declared env vars (`FrameworkEnv`) and user-supplied env vars
(`UserEnv`) on `DeployOpts` / `SandboxOpts` are merged by the driver with
`UserEnv` taking precedence.

## Capabilities

Not every backend supports every feature. Drivers declare support via:

```go
type Capabilities struct {
    Sandbox         bool // supports live bind-mount dev instances
    SecureIsolation bool // has gVisor / Kata / equivalent sandbox runtime
    Exec            bool // supports exec-ing into running instances
}
```

Callers gate on these — e.g. MCP `berth_sandbox_create` should reject a
sandbox request when `Runtime.Capabilities().Sandbox == false`.

## Registering a Driver

Each driver self-registers in an `init()`:

```go
package kubernetes

import "github.com/AmirSoleimani/openberth/apps/server/internal/runtime"

func init() {
    runtime.Register(runtime.Driver{
        Name:    "kubernetes",
        Factory: New,
    })
}

func New(cfg *config.Config) (runtime.Runtime, error) {
    // …validate cfg, open k8s client, return *Driver…
}
```

Then `main.go` blank-imports the package so its `init` runs:

```go
import _ "github.com/AmirSoleimani/openberth/apps/server/internal/runtime/kubernetes"
```

Users opt in via config:

```json
{
  "runtime": {
    "driver": "kubernetes"
  }
}
```

## Sketching Alternate Drivers

### Kubernetes

| Interface call     | Implementation hint                                                                                   |
|--------------------|-------------------------------------------------------------------------------------------------------|
| `Deploy`           | Submit a `Job` for build (mounts source, writes artifacts into a `PersistentVolumeClaim`), then apply `Deployment`+`Service`. `InstanceID` = pod UID. `Endpoint.Host` = service ClusterIP or DNS. |
| `Rebuild`          | New `Job`, then `kubectl rollout restart` or update the Deployment image to trigger a rolling swap.   |
| `RestartRuntime`   | `kubectl rollout restart deployment/<name>` — no build job.                                           |
| `StartSandbox`     | `Deployment` mounting a PVC shared with a syncer sidecar (e.g. `mutagen` or a custom tarball stream). |
| `Destroy`          | `kubectl delete deployment,service,pvc -l openberth.id=<id>`.                                         |
| `Status`           | Read `Pod.Status.Phase`; map to `runtime.Status` vocabulary.                                          |
| `Exec` / `Logs`    | `kubectl exec` / `kubectl logs`.                                                                      |
| `Port`             | Return the Service port; proxy would need to route via cluster DNS (see *Proxy* below).               |
| `Capabilities`     | `Sandbox: true` once syncer works; `SecureIsolation: true` if gVisor runtime class is installed.      |

Stays zero-Go-dep by shelling to `kubectl`, matching the project's style.
A maintainer who accepts `client-go` could swap in the typed client later.

### Firecracker

| Interface call     | Implementation hint                                                                    |
|--------------------|----------------------------------------------------------------------------------------|
| `Deploy`           | Build on host → produce a rootfs ext4 image → launch a microVM via the Firecracker API socket. `InstanceID` = VM instance ID. |
| `Rebuild`          | Build a new image, snapshot/swap VMs, tear down old.                                   |
| `RestartRuntime`   | Reboot the microVM with the new env (serial config or boot args).                      |
| `StartSandbox`     | Likely unsupported initially — set `Capabilities.Sandbox = false`. (Live mount into a microVM is hard without vsock-based sync.) |
| `Destroy`          | Kill VM; delete rootfs image.                                                          |
| `Status`           | Query Firecracker API; map InstanceInfo to `runtime.Status`.                           |
| `Exec`             | Run via a vsock agent, or expose SSH on tap device; may require `Capabilities.Exec = false` in minimal impls. |
| `Logs`             | Serial console → ring buffer; expose via `LogStream`.                                  |
| `Port`             | Tap device + iptables → host port forward.                                             |
| `Capabilities`     | `Sandbox: false` (initial), `SecureIsolation: true` (VM-level), `Exec: false` or agent-dependent. |

## Install-time Setup

Drivers can also contribute **install-time steps** via the parallel
`install.Installer` registry. When the operator runs
`berth-server install --runtime=<name>`, the installer orchestrator runs:

1. **Preflight** (universal): check root, install OS packages
2. **Driver-specific** (provided by your driver's `Steps()`): install
   backend, pull images, register kernel modules, etc.
3. **Universal infra**: Caddy, dirs, config, DB, Caddyfile, binary,
   systemd unit
4. **Universal activation**: enable services, firewall, DNS, health check

A driver that needs host provisioning adds an `installer.go`:

```go
package kubernetes

import "github.com/AmirSoleimani/openberth/apps/server/internal/install"

func init() {
    install.Register(k8sInstaller{})
}

type k8sInstaller struct{}

func (k8sInstaller) Name() string { return "kubernetes" }

func (k8sInstaller) Steps() []install.Step {
    return []install.Step{
        {
            Name:        "install_kubectl",
            Description: "Installing kubectl",
            Run:         installKubectl,
        },
        // … more driver-specific steps
    }
}

func installKubectl(ctx *install.Ctx) error {
    if out, _ := ctx.Cmd("command -v kubectl"); out != "" {
        ctx.Done("kubectl already installed")
        return nil
    }
    // … download + install …
    ctx.Done("kubectl installed")
    return nil
}
```

Drivers that assume the host is already provisioned (e.g. an external
Kubernetes cluster with credentials already present) register an empty
`Steps()` slice — or skip `install.Register` entirely, in which case
`berth-server install --runtime=<name>` reports `no installer registered
for runtime "<name>"` and the operator does the bootstrap manually.

The `Ctx` passed to each step exposes:
- `Config() *install.Config` — domain, admin key, driver name, flags
- `Cmd(cmd) (string, error)` — run shell command
- `Write(path, content, mode) error` — write file
- `Done(msg)` — emit success with a custom message
- `Warn(msg, detail)` — emit non-fatal warning

See `runtime/docker/installer.go` for the reference implementation.

## What's Not Covered by the Driver

- **Framework detection** (`internal/framework/`) is runtime-agnostic — it
  produces shell scripts against the contract above, regardless of backend.
- **Proxy routing** (`internal/proxy/`) currently assumes `localhost:port`
  (Docker-friendly). The `Endpoint.Host` field is reserved for drivers that
  route via non-loopback addresses (K8s service DNS, microVM tap IPs);
  proxy refactoring to consume it is a separate task.
- **Secrets** (`internal/secret/`) are decrypted server-side and injected
  into `DeployOpts.UserEnv` before the driver sees anything. No driver
  handles ciphertext.

## Reference Implementation

`runtime/docker/` is the canonical driver, split across:

- `driver.go` — struct, `New` factory, `init()` registration, helpers
  (gVisor probe, port allocator, exec helpers, framework-info adapters)
- `build.go` — `Deploy`, `Rebuild`, `RestartRuntime`, `runBuild`, `startRuntime`
- `sandbox.go` — `StartSandbox`
- `static.go` — `createStatic`, `rebuildStatic`, `createStaticSandbox`
  (Caddy-based static file serving)
- `lifecycle.go` — `Exec`, `Destroy`, `Logs`, `LogStream`, `Status`,
  `Restart`, `Port`, and the Docker-state → `runtime.Status` mapping

Copy this layout when implementing a new driver — each file stays
single-concern and manageable (<300 lines).
