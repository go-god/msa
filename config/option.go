package config

// Option for configSetting
type Option func(*ConfigOption)

// WithConfigDir set config dir
func WithConfigDir(dir string) Option {
	return func(c *ConfigOption) {
		c.configDir = dir
	}
}

// WithConfigFile set config filename
func WithConfigFile(file string) Option {
	return func(c *ConfigOption) {
		c.configFile = file
	}
}
