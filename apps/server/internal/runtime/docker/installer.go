package docker

import (
	"fmt"

	"github.com/AmirSoleimani/openberth/apps/server/internal/install"
)

// Docker-specific install steps. Contributed to the install orchestrator
// via install.Register from init(). When cfg.Runtime.Driver = "docker"
// (the default), these steps run in Phase 2 of the install sequence,
// between the universal preflight and universal infra phases.

func init() {
	install.Register(dockerInstaller{})
}

type dockerInstaller struct{}

func (dockerInstaller) Name() string { return "docker" }

func (dockerInstaller) Steps() []install.Step {
	return []install.Step{
		{Name: "install_docker", Description: "Installing Docker", Run: installDocker},
		{Name: "install_gvisor", Description: "Installing gVisor sandbox runtime", Run: installGVisor},
		{Name: "test_gvisor", Description: "Testing gVisor runtime", Run: testGVisor},
		{Name: "pull_images", Description: "Pulling base Docker images", Run: pullImages},
		{Name: "create_volumes", Description: "Creating Docker volumes", Run: createVolumes},
	}
}

func installDocker(ctx *install.Ctx) error {
	if out, _ := ctx.Cmd("command -v docker"); out != "" {
		ctx.Done("Docker already installed")
		return nil
	}

	cmds := []string{
		"install -m 0755 -d /etc/apt/keyrings",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null",
		"chmod a+r /etc/apt/keyrings/docker.gpg",
		`echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list`,
		"apt-get update -qq",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin >/dev/null 2>&1",
		"systemctl enable --now docker",
	}

	for _, cmd := range cmds {
		if _, err := ctx.Cmd(cmd); err != nil {
			return fmt.Errorf("docker install: %w", err)
		}
	}
	ctx.Done("Docker installed")
	return nil
}

func installGVisor(ctx *install.Ctx) error {
	if out, _ := ctx.Cmd("command -v runsc"); out != "" {
		ctx.Done("gVisor already installed")
		return nil
	}

	dl := `ARCH=$(uname -m) && URL="https://storage.googleapis.com/gvisor/releases/release/latest/${ARCH}" && curl -fsSL "${URL}/runsc" -o /usr/local/bin/runsc && curl -fsSL "${URL}/containerd-shim-runsc-v1" -o /usr/local/bin/containerd-shim-runsc-v1 && chmod +x /usr/local/bin/runsc /usr/local/bin/containerd-shim-runsc-v1`
	if _, err := ctx.Cmd(dl); err != nil {
		return fmt.Errorf("gvisor install: %w", err)
	}

	if err := ctx.Write("/etc/docker/daemon.json", daemonJSONTemplate, 0644); err != nil {
		return fmt.Errorf("write daemon.json: %w", err)
	}

	if _, err := ctx.Cmd("systemctl restart docker"); err != nil {
		return fmt.Errorf("restart docker: %w", err)
	}

	ctx.Done("gVisor installed and registered")
	return nil
}

func testGVisor(ctx *install.Ctx) error {
	if _, err := ctx.Cmd("docker run --rm --runtime=runsc hello-world >/dev/null 2>&1"); err != nil {
		ctx.Warn("gVisor test failed", "will fall back to runc — check KVM/CPU support")
		return nil
	}
	ctx.Done("gVisor runtime verified")
	return nil
}

func pullImages(ctx *install.Ctx) error {
	if _, err := ctx.Cmd("docker pull node:20-slim -q && docker pull caddy:2-alpine -q"); err != nil {
		return fmt.Errorf("pull images: %w", err)
	}
	ctx.Done("Base images pulled")
	return nil
}

func createVolumes(ctx *install.Ctx) error {
	if _, err := ctx.Cmd("docker volume create openberth-npm-cache >/dev/null 2>&1 || true"); err != nil {
		return fmt.Errorf("create volume: %w", err)
	}
	ctx.Done("Docker volumes created")
	return nil
}

// daemonJSONTemplate is the /etc/docker/daemon.json content that registers
// the runsc (gVisor) runtime. Written by installGVisor.
const daemonJSONTemplate = `{
    "runtimes": {
        "runsc": {
            "path": "/usr/local/bin/runsc",
            "runtimeArgs": ["--network=sandbox", "--platform=systrap"]
        }
    }
}`
