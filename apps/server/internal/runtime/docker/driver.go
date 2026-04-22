// Package docker is the reference Runtime driver. It shells out to the
// local `docker` CLI to build, run, update, and destroy instances. gVisor
// (runsc) is used for security isolation when available; otherwise the
// driver falls back to the default OCI runtime (runc).
//
// This driver is registered under the name "docker" and is the default
// when cfg.Runtime.Driver is empty.
package docker

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/AmirSoleimani/openberth/apps/server/internal/config"
	"github.com/AmirSoleimani/openberth/apps/server/internal/framework"
	"github.com/AmirSoleimani/openberth/apps/server/internal/runtime"
)

// Driver is the Runtime implementation backed by the Docker engine.
type Driver struct {
	cfg         *config.Config
	gvisorReady bool
}

func init() {
	runtime.Register(runtime.Driver{
		Name:    "docker",
		Factory: New,
	})
}

// New constructs a Docker Driver. It probes the local Docker daemon for
// gVisor support so callers can declare secure isolation via Capabilities.
func New(cfg *config.Config) (runtime.Runtime, error) {
	d := &Driver{cfg: cfg}
	d.gvisorReady = d.checkGVisor()
	return d, nil
}

// Capabilities reports this driver's optional features.
func (d *Driver) Capabilities() runtime.Capabilities {
	return runtime.Capabilities{
		Sandbox:         true,
		SecureIsolation: d.gvisorReady,
		Exec:            true,
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────

func (d *Driver) checkGVisor() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--runtime=runsc", "hello-world")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (d *Driver) findPort() (int, error) {
	for i := 0; i < 100; i++ {
		port := 10000 + rand.Intn(50000)
		out, _ := execCmd("ss", "-tlnp")
		if !strings.Contains(out, fmt.Sprintf(":%d ", port)) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("could not find available port")
}

// volumeForDeploy returns a unique named volume identifier for a deploy.
// The timestamp suffix lets blue-green rebuilds keep the old volume alive
// for rollback until the new runtime is verified healthy.
func volumeForDeploy(deployID string) string {
	return fmt.Sprintf("sc-ws-%s-%d", deployID, time.Now().UnixMilli())
}

// currentVolume finds the named volume currently mounted at /app for the
// running container. Used by blue-green rebuild and secret-rotation paths.
func (d *Driver) currentVolume(deployID string) string {
	out, err := execCmd("docker", "inspect", "-f",
		`{{range .Mounts}}{{if eq .Destination "/app"}}{{.Name}}{{end}}{{end}}`,
		"sc-"+deployID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// runtimeImage returns the runtime-phase image, falling back to Image if
// the build and runtime images are the same (interpreted languages).
func runtimeImage(opts runtime.DeployOpts) string {
	if opts.RunImage != "" {
		return opts.RunImage
	}
	return opts.Image
}

func execCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func execCmdTimeout(name string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("timed out after %s", timeout)
	}
	return string(out), err
}

// fwInfoFromOpts adapts runtime.DeployOpts into the framework provider's
// input shape. Kept as a driver helper so the runtime package stays free
// of the framework dependency.
func fwInfoFromOpts(opts runtime.DeployOpts) *framework.FrameworkInfo {
	return &framework.FrameworkInfo{
		Framework:  opts.Framework,
		Language:   opts.Language,
		BuildCmd:   opts.BuildCmd,
		StartCmd:   opts.StartCmd,
		InstallCmd: opts.InstallCmd,
		Port:       opts.Port,
		Image:      opts.Image,
		RunImage:   opts.RunImage,
		CacheDir:   opts.CacheDir,
		Env:        opts.FrameworkEnv,
	}
}

func fwInfoFromSandboxOpts(opts runtime.SandboxOpts) *framework.FrameworkInfo {
	return &framework.FrameworkInfo{
		Framework:  opts.Framework,
		Language:   opts.Language,
		DevCmd:     opts.DevCmd,
		InstallCmd: opts.InstallCmd,
		Port:       opts.Port,
		Image:      opts.Image,
		Env:        opts.FrameworkEnv,
	}
}

// makeResult is a small constructor that all Deploy/Rebuild/Restart paths
// use to produce a runtime.Result. Centralising it keeps the Endpoint /
// Metadata shape consistent and easy to evolve.
func (d *Driver) makeResult(cid, containerName string, hostPort int) *runtime.Result {
	return &runtime.Result{
		InstanceID: cid,
		Endpoint: runtime.Endpoint{
			Host: "127.0.0.1",
			Port: hostPort,
		},
		Isolated: d.gvisorReady,
		Metadata: map[string]string{"containerName": containerName},
	}
}
