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
	stopCh           chan struct{}       // stop chan,if you call ActiveStop() application will exit

	// config provider optional
	configInterface config.ConfigInterface    // config read interface
	configProviders []provides.ConfigProvider // all config providers
}

// engine default engine
var engine *Engine

// Start create an engine and run application.
func Start(opts ...Option) {
	engine = New(opts...)
	engine.Start()
}

// ActiveStop if receive active exit signal,the application will exit
func ActiveStop() {
	engine.ActiveStop()
}

func defaultInjector() gdi.Injector {
	return factory.CreateDI(factory.FbInject)
}

func defaultConfig() config.ConfigInterface {
	return config.New()
}

// New create an application engine
func New(opts ...Option) *Engine {
	e := &Engine{
		gracefulWait:     5 * time.Second,
		signal:           make(chan os.Signal, 1),
		interruptSignals: InterruptSignals,
		stopCh:           make(chan struct{}, 1),
		injector:         defaultInjector(),
		configInterface:  defaultConfig(),
	}

	for _, o := range opts {
		o(e)
	}

	return e
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

	for _, p := range e.configProviders { // config providers
		provides.Register(p.Provide(e.configInterface))
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

	// before invoke init action
	for _, val := range e.injectValues {
		if initStream, ok := val.Value.(initInterface); ok {
			err = initStream.Init()
			if err != nil {
				panic("init error: " + err.Error())
			}
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

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// recv signal to exit main goroutine
	signal.Notify(e.signal, e.interruptSignals...)
	// Block until we receive our signal.
	select {
	case sig := <-e.signal:
		log.Println("receive exit signal: ", sig.String())
	case <-e.stopCh:
		log.Println("receive active exit signal")
	}

	e.stop()
}

// stop stop application
func (e *Engine) stop() {
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), e.gracefulWait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	// Optionally, you could run srv.Shutdown in a goroutine and block on
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

// ActiveStop if receive active exit signal,the application will exit
func (e *Engine) ActiveStop() {
	close(e.stopCh)
}
