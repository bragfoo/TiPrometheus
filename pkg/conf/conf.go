package conf

// AgentConf is the agent configuration type.
type AgentConf struct {
	PDHost       string
	TimeInterval int
	AdapterListen string
}

// RunTimeInfo contains the active configuration.
var RunTimeInfo AgentConf

// RunTimeMap contains the full configuration file.
var RunTimeMap map[string]AgentConf

// DefaultRunTimeName specifies the name of the default TOML section
// to be loaded when nothing is configured.
const DefaultRunTimeName = "default"
