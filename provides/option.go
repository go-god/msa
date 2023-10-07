package provides

// Option providerOption functional option
type Option func(o providerOption)
type providerOption struct {
	name  string // provider name
	group string // provider group
}

// WithProviderName set provider name
func WithProviderName(name string) Option {
	return func(o providerOption) {
		o.name = name
	}
}

// WithProviderGroup set provider group
func WithProviderGroup(group string) Option {
	return func(o providerOption) {
		o.group = group
	}
}
