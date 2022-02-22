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

// initInterface init interface
type initInterface interface {
	Init() error
}

type stopInterface interface {
	Stop(ctx context.Context)
}

type startInterface interface {
	Start() error
}

// Engine application engine
type Engine struct {
	interruptSignals []os.Signal         // interrupt signals
	gracefulWait     time.Duration       // graceful exit time
	signal           chan os.Signal      // recv interrupt signals
	injectValues     []*gdi.Object       // inject objects
	injector         gdi.Injector        // dip inject interface
	injectType       factory.InjectType  // inject type as: factory.FbInject or factory.DigInject
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

// inject values stop action
func (e *Engine) shutdown(ctx context.Context) {
	for _, val := range e.injectValues {
		if stopStream, ok := val.Value.(stopInterface); ok {
			stopStream.Stop(ctx)
		}
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

// Start start application
func (e *Engine) Start() {
	// load all provides
	e.loadProvides()

	var err error
	// init inject objects
	if len(e.injectValues) > 0 {
		err = e.injector.Provide(e.injectValues...)
		if err != nil {
			panic("provide inject objects error: " + err.Error())
		}
	}

	// invoke objects
	if len(e.invokeFunc) > 0 {
		err = e.injector.Invoke(e.invokeFunc...)
	} else {
		err = e.injector.Invoke()
	}

	if err != nil {
		panic("inject invoke error: " + err.Error())
	}

	// after invoke init action
	// perform some init operations after the binding is performed.
	for _, val := range e.injectValues {
		if initStream, ok := val.Value.(initInterface); ok {
			err = initStream.Init()
			if err != nil {
				panic("init error: " + err.Error())
			}
		}
	}

	for _, val := range e.injectValues {
		if startStream, ok := val.Value.(startInterface); ok {
			err = startStream.Start()
			if err != nil {
				panic("start error: " + err.Error())
			}
		}
	}

	log.Println("msa started successfully")

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// recv signal to exit main goroutine
	signal.Notify(e.signal, e.interruptSignals...)
	// Block until we receive our signal.
	select {
	case sig := <-e.signal:
		log.Println("receive exit signal: ", sig.String())
		e.gracefulStop()
	case <-e.stopCh:
	}
}

// gracefulStop stop application
func (e *Engine) gracefulStop() {
	defer log.Println("msa exit successfully")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), e.gracefulWait)
	defer cancel()

	// Doesn't block if no service run, but will otherwise wait
	// until the timeout deadline.
	// Optionally, you could run it in a goroutine and block on
	// if your application should wait for other services
	// to finalize based on context cancellation.
	done := make(chan struct{}, 1)
	go func() {
		defer close(done)

		e.shutdown(ctx)
	}()

	<-done
	<-ctx.Done()

	log.Println("server shutting down")
}

// Stop if receive active exit signal,the application will exit
func (e *Engine) Stop() {
	log.Println("receive stop action signal")
	e.gracefulStop()
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
