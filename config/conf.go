package config

// ConfigInterface
type ConfigInterface interface {
	// Load load config
	Load(opts ...Option) error
	// IsSet is set value
	IsSet(key string) bool
	// GetValue get key to obj,obj must be a pointer
	GetValue(key string, obj interface{}) error
}
