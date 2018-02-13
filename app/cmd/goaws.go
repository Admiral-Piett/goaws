package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/p4tin/goaws/app/conf"
	"github.com/p4tin/goaws/app/gosqs"
	"github.com/p4tin/goaws/app/router"
)

func main() {
	var filename string
	var debug bool
	flag.StringVar(&filename, "config", "", "config file location + name")
	flag.BoolVar(&debug, "debug", false, "debug log level (default Warning)")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	env := "Local"
	if flag.NArg() > 0 {
		env = flag.Arg(0)
	}

	portNumbers := conf.LoadYamlConfig(filename, env)

	r := router.New()

	quit := make(chan struct{}, 0)
	go gosqs.PeriodicTasks(1*time.Second, quit)

	if len(portNumbers) == 1 {
		log.Warnf("GoAws listening on: 0.0.0.0:%s", portNumbers[0])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
		log.Fatal(err)
	} else if len(portNumbers) == 2 {
		go func() {
			log.Warnf("GoAws listening on: 0.0.0.0:%s", portNumbers[0])
			err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
			log.Fatal(err)
		}()
		log.Warnf("GoAws listening on: 0.0.0.0:%s", portNumbers[1])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[1], r)
		log.Fatal(err)
	} else {
		log.Fatal("Not enough or too many ports defined to start GoAws.")
	}
}
