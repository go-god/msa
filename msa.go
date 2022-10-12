package msa

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-god/gdi"
	"github.com/go-god/gdi/factory"
	"github.com/go-god/msa/config"
	"github.com/go-god/msa/provides"
)

// initializer init interface
type initializer interface {
	Init() error
}

// starter start interface
type starter interface {
	Start() error
}

// stoppable stop interface
type stoppable interface {
	Stop()
}

// Engine application engine
type Engine struct {
	interruptSignals []os.Signal         // interrupt signals
	gracefulWait     time.Duration       // graceful exit time
	signal           chan os.Signal      // recv interrupt signals
	injectValues     []*gdi.Object       // inject objects
	injector         gdi.Injector        // dip inject interface
	invokeFunc       []interface{}       // invoke func
	providers        []provides.Provider // all provides
	stopCh           chan struct{}       // stop chan,if you call Stop() application will exit

	// config provider these are optional parameters
	configDir       string                  // config dirname
	configFile      string                  // config file
	configInterface config.ConfigInterface  // config read interface
	configProvider  provides.ConfigProvider // all provides.ConfigProvider
}

// engine default engine
var engine *Engine

// Start create an engine and run application.
func Start(opts ...Option) {
	engine = New(opts...)
	engine.Start()
}

// Stop if receive active exit signal,the application will exit
func Stop() {
	engine.Stop()
}

// LoadConf get key from configInterface,obj must be a pointer
func LoadConf(key string, obj interface{}) error {
	return engine.LoadConf(key, obj)
}

// IsSet check configInterface is set key
func IsSet(key string) bool {
	return engine.IsSet(key)
}

func defaultInjector() gdi.Injector {
	return factory.CreateDI(factory.FbInject)
}

func defaultConfig() config.ConfigInterface {
	return config.New()
}

// New create an application for msa engine
func New(opts ...Option) *Engine {
	e := &Engine{
		gracefulWait:     5 * time.Second,
		signal:           make(chan os.Signal, 1),
		interruptSignals: InterruptSignals,
		stopCh:           make(chan struct{}, 1),
		injector:         defaultInjector(),
	}

	for _, o := range opts {
		o(e)
	}

	// if opts has no ConfigInterface will use it
	if e.configInterface == nil {
		e.configInterface = defaultConfig()
	}

	// if the configuration file directory and file, regenerate a config interface.
	e.resetConfInterface()

	return e
}

// Start run app
func (e *Engine) Start() {
	// load all provides
	e.loadProvides()

	// invoke inject objects
	e.invokeInjects()

	// wait exit signal
	e.waitExitSignal()

	// graceful stop
	e.gracefulStop()
}

func (e *Engine) resetConfInterface() {
	var confOptions []config.Option
	if e.configDir != "" {
		confOptions = append(confOptions, config.WithConfigDir(e.configDir))
	}
	if e.configFile != "" {
		confOptions = append(confOptions, config.WithConfigFile(e.configFile))
	}
	if len(confOptions) > 0 {
		e.configInterface = config.New(confOptions...)
	}
}

// loadProvides load providers and config inject providers
func (e *Engine) loadProvides() {
	for _, p := range e.providers {
		provides.Register(p)
	}

	if e.configProvider != nil {
		// register all providers from configProvider
		configProviders := e.configProvider.Provide(e.configInterface)
		for _, p := range configProviders {
			provides.Register(p)
		}
	}

	if provideObjects := provides.ProvideObjects(); len(provideObjects) > 0 {
		e.injectValues = append(e.injectValues, provideObjects...)
	}
}

func (e *Engine) waitExitSignal() {
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// receive signal to exit main goroutine
	// Block until we receive our signal.
	signal.Notify(e.signal, e.interruptSignals...)
	select {
	case sig := <-e.signal:
		signal.Stop(e.signal)
		log.Println("receive exit signal: ", sig.String())
	case <-e.stopCh:
	}
}

func (e *Engine) invokeInjects() {
	// init inject objects
	if len(e.injectValues) > 0 {
		if err := e.injector.Provide(e.injectValues...); err != nil {
			panic("provide inject objects error: " + err.Error())
		}
	}

	// invoke objects
	if err := e.injector.Invoke(e.invokeFunc...); err != nil {
		panic("inject invoke error: " + err.Error())
	}

	// after invoke init action
	// perform some init operations after the binding is performed.
	for _, val := range e.injectValues {
		if initStream, ok := val.Value.(initializer); ok {
			if err := initStream.Init(); err != nil {
				panic("init error: " + err.Error())
			}
		}
	}

	for _, val := range e.injectValues {
		if startStream, ok := val.Value.(starter); ok {
			if err := startStream.Start(); err != nil {
				panic("start error: " + err.Error())
			}
		}
	}

	log.Println("msa started successfully")
}

// gracefulStop stop application
func (e *Engine) gracefulStop() {
	defer log.Println("msa exit successfully")

	for _, val := range e.injectValues {
		if s, ok := val.Value.(stoppable); ok {
			s.Stop()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.gracefulWait)
	defer cancel()
	<-ctx.Done()
}

// Stop if receive active exit signal,the application will exit
func (e *Engine) Stop() {
	log.Println("receive stop action signal")
	close(e.stopCh)
}

// LoadConf get key from configInterface,obj must be a pointer
func (e *Engine) LoadConf(key string, obj interface{}) error {
	return e.configInterface.GetValue(key, obj)
}

// IsSet configInterface is set key
func (e *Engine) IsSet(key string) bool {
	return e.configInterface.IsSet(key)
}
