package msa

import (
	"os"
	"time"

	"github.com/go-god/gdi"
	"github.com/go-god/gdi/factory"

	"github.com/go-god/msa/config"
	"github.com/go-god/msa/logger"
	"github.com/go-god/msa/provides"
)

// Option engine option
type Option func(e *Engine)

// WithGracefulWait set engine gracefulWait.
func WithGracefulWait(t time.Duration) Option {
	return func(e *Engine) {
		e.gracefulWait = t
	}
}

// WithInjector set injector
// inject type as: factory.FbInject or factory.DigInject
func WithInjector(injectType factory.InjectType) Option {
	return func(e *Engine) {
		e.injector = factory.CreateDI(injectType)
	}
}

// WithInjectValues set inject object
func WithInjectValues(objects ...*gdi.Object) Option {
	return func(e *Engine) {
		e.injectValues = append(e.injectValues, objects...)
	}
}

// WithInterruptSignals set engine interruptSignals
func WithInterruptSignals(signals ...os.Signal) Option {
	return func(e *Engine) {
		e.interruptSignals = append(e.interruptSignals, signals...)
	}
}

// WithConfigInterface set config read interface
func WithConfigInterface(c config.ConfigInterface) Option {
	return func(e *Engine) {
		e.configInterface = c
	}
}

// WithProviders add providers
func WithProviders(provides ...provides.Provider) Option {
	return func(e *Engine) {
		e.providers = append(e.providers, provides...)
	}
}

// WithConfigProvider add config providers
func WithConfigProvider(configProvider provides.ConfigProvider) Option {
	return func(e *Engine) {
		e.configProvider = configProvider
	}
}

// WithLogger logger config
func WithLogger(opts ...logger.Option) Option {
	return func(e *Engine) {
		logger.Default(opts...)
	}
}
