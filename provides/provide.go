package provides

import (
	"github.com/go-god/gdi"

	"github.com/go-god/msa/config"
)

// Provider config provider interface
type Provider interface {
	Provide() *gdi.Object
}

// Option providerOption functional option
type Option func(o providerOption)
type providerOption struct {
	name  string // provider name
	group string // // provider group
}

// ConfigProvider config provider
type ConfigProvider interface {
	Provide(c config.ConfigInterface) []Provider
}

var provideObjects = make([]*gdi.Object, 0, 20)

// Register register Provider
func Register(p Provider, opts ...Option) {
	obj := p.Provide()
	providerOpt := providerOption{}
	for _, o := range opts {
		o(providerOpt)
	}

	if providerOpt.name != "" {
		obj.Name = providerOpt.name
	}

	if providerOpt.group != "" {
		obj.Group = providerOpt.group
	}

	provideObjects = append(provideObjects, obj)
}

// ProvideObjects return provideObjects
func ProvideObjects() []*gdi.Object {
	return provideObjects
}
