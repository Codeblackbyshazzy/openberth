package install

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// Installer contributes runtime-driver-specific install steps. Drivers that
// need host provisioning (install Docker, pull images, configure a kernel,
// lay down helm charts) implement this and call Register in an init
// function. Drivers with no install needs — e.g. a Kubernetes driver that
// assumes an existing cluster + kubeconfig — can skip this entirely.
//
// The install orchestrator runs steps in four phases:
//  1. Universal preflight  (check root, install OS packages)
//  2. Driver-specific      (these steps)
//  3. Universal infra      (Caddy, dirs, config, DB, Caddyfile, binary, admin, systemd)
//  4. Universal activation (enable services, firewall, DNS, health, summary)
//
// Driver steps therefore run after OS packages are present but before
// Caddy and OpenBerth itself are set up.
type Installer interface {
	Name() string   // must match the runtime driver's Name
	Steps() []Step
}

// Step is one discrete install action. Name is the stable ID used in
// progress events. Description is the verb-phrase shown alongside the
// spinner ("Installing Docker…"). Run performs the work.
type Step struct {
	Name        string
	Description string
	Run         func(*Ctx) error
}

// Ctx is the context passed to every Step.Run. It exposes the install
// config, shell/file helpers, and event emitters. Steps never touch the
// provisioner directly — Ctx is the public surface for driver steps.
type Ctx struct {
	prov     *provisioner
	name     string
	progress int
	emitted  bool // set when Done or Warn is called, suppressing the default Completed emit
}

// Config returns the install-time configuration (domain, admin key,
// runtime driver, etc).
func (c *Ctx) Config() *Config { return c.prov.cfg }

// Cmd runs a shell command via /bin/bash -c and returns trimmed combined output.
func (c *Ctx) Cmd(command string) (string, error) {
	return runCmd(command)
}

// Write writes content to a local file with the given permissions.
func (c *Ctx) Write(path, content string, mode os.FileMode) error {
	return writeFile(path, content, mode)
}

// Done emits StepCompleted with a custom message. Use when the default
// Step.Description doesn't capture the outcome (e.g. "Docker already
// installed" vs "Docker installed"). After Done, the orchestrator skips
// its default Completed emit for this step.
func (c *Ctx) Done(msg string) {
	if c.emitted {
		return
	}
	c.prov.emit(c.name, StepCompleted, msg, "", c.progress)
	c.emitted = true
}

// Warn emits StepWarning for a non-fatal issue and marks the step done.
// The step should return nil after calling Warn so the install run
// continues.
func (c *Ctx) Warn(msg, detail string) {
	if c.emitted {
		return
	}
	c.prov.emit(c.name, StepWarning, msg, detail, c.progress)
	c.emitted = true
}

var (
	installersMu sync.RWMutex
	installers   = map[string]Installer{}
)

// Register adds an Installer to the registry. Intended for use from init()
// functions in driver packages. Panics on duplicate names so contributors
// catch accidental collisions at startup.
func Register(i Installer) {
	if i.Name() == "" {
		panic("install.Register: empty installer name")
	}
	installersMu.Lock()
	defer installersMu.Unlock()
	if _, dup := installers[i.Name()]; dup {
		panic("install.Register: duplicate installer " + i.Name())
	}
	installers[i.Name()] = i
}

// GetInstaller looks up a registered installer by driver name. Returns a
// helpful error listing what IS registered when the lookup misses.
func GetInstaller(name string) (Installer, error) {
	installersMu.RLock()
	defer installersMu.RUnlock()
	i, ok := installers[name]
	if !ok {
		return nil, fmt.Errorf("no installer registered for runtime %q (available: %v)", name, installerNames())
	}
	return i, nil
}

func installerNames() []string {
	names := make([]string, 0, len(installers))
	for n := range installers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
