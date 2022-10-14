package main

import (
	"log"
	"time"

	"github.com/go-god/gdi"
	"github.com/go-god/msa"
)

// Service svc
type Service struct {
	AppName string `json:"app_name"`
	AppEnv  string `json:"app_env"`
}

// App application
type App struct {
	Service *Service `inject:""`
}

// Init init action
func (a *App) Init() error {
	err := msa.LoadConf("service", a.Service)
	if err != nil {
		return err
	}

	log.Println("app_name: ", a.Service.AppName)
	log.Println("app_env", a.Service.AppEnv)
	return nil
}

// Start app start action
func (a *App) Start() error {
	log.Println("app startup successful")
	return nil
}

func main() {
	app := &App{}
	opts := []msa.Option{
		msa.WithGracefulWait(1 * time.Second),
		msa.WithInjectValues(&gdi.Object{
			Value: app,
		}),
		msa.WithInjectValues(&gdi.Object{
			Value: &Service{},
		}),
	}

	// the first way
	// engine := msa.New(opts...)
	// engine.LoadConf("service", app.Service)
	// log.Println("app_name: ", app.Service.AppName)
	// log.Println("app_env", app.Service.AppEnv)
	// engine.Start()

	// the second way
	msa.Start(opts...)
}

/*
 % ./inject_demo
2022/10/14 22:21:49 app_name:  demo
2022/10/14 22:21:49 app_env local
2022/10/14 22:21:49 app startup successful
2022/10/14 22:21:49 msa started successfully
^C2022/10/14 22:23:43 receive exit signal:  interrupt
2022/10/14 22:23:44 msa exit successfully
*/
