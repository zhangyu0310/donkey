package config

import "sync/atomic"

type Config struct {
	DbType         string
	Host           string
	Port           int
	User           string
	Pass           string
	Database       string
	InsertRows     uint64
	FrontSQL       string
	PostSQL        string
	UniqueSyntax   string
	RoutineNum     uint
	InsertData     bool
	CheckData      bool
	InsertPackage  uint
	ExtraColumnNum uint
	InsertDelay    int64
	TimeConsume    bool
}

var globalCfg atomic.Value

// InitializeConfig initialize the global config handler.
func InitializeConfig(enforceCmdArgs func(*Config)) {
	cfg := Config{}
	// Use command config cover config file.
	enforceCmdArgs(&cfg)
	StoreGlobalConfig(&cfg)
}

// GetGlobalConfig returns the global configuration for this server.
// It should store configuration from command line and configuration file.
// Other parts of the system can read the global configuration use this function.
func GetGlobalConfig() *Config {
	return globalCfg.Load().(*Config)
}

// StoreGlobalConfig stores a new config to the globalConf. It mostly uses in the test to avoid some data races.
func StoreGlobalConfig(config *Config) {
	globalCfg.Store(config)
}
