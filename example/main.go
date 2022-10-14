package main

import (
	"log"
	"time"

	"github.com/go-god/msa"
	"github.com/go-god/msa/config"
)

func main() {
	// msa.Start()

	// go msa.Start()
	// time.Sleep(5 * time.Second)
	// msa.Stop()

	engine := msa.New(
		msa.WithGracefulWait(5*time.Second),
		msa.WithConfigInterface(config.New(config.WithConfigFile("test.yaml"))),
	)
	var appName string
	engine.LoadConf("app_name", &appName)
	log.Println("app_name: ", appName)
	engine.Start()
}

/*
2022/10/14 21:37:10 app_name:  demo
2022/10/14 21:37:10 msa started successfully
^C2022/10/14 21:38:09 receive exit signal:  interrupt
2022/10/14 21:38:14 msa exit successfully
*/
