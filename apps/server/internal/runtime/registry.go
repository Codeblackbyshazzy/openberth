package runtime

import (
	"fmt"
	"sort"
	"sync"

	"github.com/AmirSoleimani/openberth/apps/server/internal/config"
)

// Driver is a registered runtime implementation. Drivers self-register
// in init() functions so main.go only needs to blank-import them.
type Driver struct {
	Name    string
	Factory func(cfg *config.Config) (Runtime, error)
}

var (
	driversMu sync.RWMutex
	drivers   = map[string]Driver{}
)

// Register adds a Driver to the registry. Intended for use from package
// init() functions. Panics on duplicate names so contributors catch
// accidental collisions at startup rather than silently shadowing a
// built-in driver.
func Register(d Driver) {
	if d.Name == "" {
		panic("runtime.Register: empty driver name")
	}
	if d.Factory == nil {
		panic("runtime.Register: nil Factory for driver " + d.Name)
	}
	driversMu.Lock()
	defer driversMu.Unlock()
	if _, dup := drivers[d.Name]; dup {
		panic("runtime.Register: duplicate driver " + d.Name)
	}
	drivers[d.Name] = d
}

// Load instantiates the driver named in cfg.Runtime.Driver. An empty
// driver name defaults to "docker" for backwards compatibility with
// existing installs whose config.json predates the Runtime block.
func Load(cfg *config.Config) (Runtime, error) {
	name := cfg.Runtime.Driver
	if name == "" {
		name = "docker"
	}
	driversMu.RLock()
	d, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown runtime driver %q (available: %v)", name, registeredNames())
	}
	return d.Factory(cfg)
}

func registeredNames() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	names := make([]string, 0, len(drivers))
	for n := range drivers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
