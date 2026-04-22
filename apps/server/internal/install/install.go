package install

import "fmt"

const dataDir = "/var/lib/openberth"

// StepStatus represents the current state of a provisioning step.
type StepStatus string

const (
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepWarning   StepStatus = "warning"
	StepFailed    StepStatus = "failed"
)

// Event represents a state change during provisioning.
type Event struct {
	Step     string
	Status   StepStatus
	Message  string
	Detail   string
	Progress int
	Total    int
}

// EventHandler is called for every state change during provisioning.
type EventHandler func(Event)

// Config holds all configuration for a local provisioning run.
type Config struct {
	Domain          string
	AdminKey        string
	AdminPassword   string
	Driver          string // runtime driver name; empty defaults to "docker"
	CloudflareProxy bool
	Insecure        bool
	WebDisabled     bool
	MaxDeploys      int
	DefaultTTL      int
}

func (c *Config) setDefaults() {
	if c.MaxDeploys == 0 {
		c.MaxDeploys = 10
	}
	if c.DefaultTTL == 0 {
		c.DefaultTTL = 72
	}
	if c.AdminKey == "" {
		c.AdminKey = generateKey()
	}
	if c.AdminPassword == "" {
		c.AdminPassword = generatePassword()
	}
	if c.Driver == "" {
		c.Driver = "docker"
	}
}

func (c *Config) validate() error {
	if c.Domain == "" {
		return fmt.Errorf("--domain is required")
	}
	if c.Insecure && c.CloudflareProxy {
		return fmt.Errorf("--insecure and --cloudflare are mutually exclusive")
	}
	return nil
}

// provisioner runs the provisioning sequence locally. The step list is
// composed at run time from four phases — see runAll. Driver-specific
// steps are contributed by the runtime driver matching cfg.Driver.
type provisioner struct {
	cfg     *Config
	onEvent EventHandler
	total   int // set at the start of runAll for progress emits
}

func (p *provisioner) emit(step string, status StepStatus, msg, detail string, progress int) {
	if p.onEvent != nil {
		p.onEvent(Event{
			Step:     step,
			Status:   status,
			Message:  msg,
			Detail:   detail,
			Progress: progress,
			Total:    p.total,
		})
	}
}

// runAll executes the four-phase install sequence. Phase 2 (driver-specific)
// is contributed by the registered Installer matching cfg.Driver; other
// phases are universal.
func (p *provisioner) runAll() error {
	inst, err := GetInstaller(p.cfg.Driver)
	if err != nil {
		return err
	}

	steps := make([]Step, 0, 20)
	steps = append(steps, preflightSteps()...)
	steps = append(steps, inst.Steps()...)
	steps = append(steps, infraSteps()...)
	steps = append(steps, activationSteps()...)

	p.total = len(steps)

	for i, s := range steps {
		stepNum := i + 1
		p.emit(s.Name, StepRunning, s.Description, "", stepNum)

		ctx := &Ctx{prov: p, name: s.Name, progress: stepNum}
		runErr := s.Run(ctx)

		if runErr != nil {
			p.emit(s.Name, StepFailed, s.Description, runErr.Error(), stepNum)
			return fmt.Errorf("step %d/%d (%s): %w", stepNum, p.total, s.Name, runErr)
		}

		// If the step didn't call Done/Warn, emit the default Completed.
		if !ctx.emitted {
			p.emit(s.Name, StepCompleted, s.Description, "", stepNum)
		}
	}

	return nil
}
