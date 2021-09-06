package provides

import (
	"github.com/go-god/gdi"
	"github.com/go-god/msa/config"
)

// Provider config provider interface
type Provider interface {
	Provide() *gdi.Object
	Name() string   // provider name
	Group() string  // provider group name
	String() string // provider string
}

// ConfigProvider config provider
type ConfigProvider interface {
	Provide(c config.ConfigInterface) Provider
}

var provideObjects = make([]*gdi.Object, 0, 20)

// Register register Provider
func Register(p Provider) {
	obj := &gdi.Object{
		Value: p.Provide(),
	}

	if name := p.Name(); name != "" {
		obj.Name = name
	}

	if group := p.Group(); group != "" {
		obj.Group = group
	}

	provideObjects = append(provideObjects, obj)
}

// ProvideObjects return provideObjects
func ProvideObjects() []*gdi.Object {
	return provideObjects
}
