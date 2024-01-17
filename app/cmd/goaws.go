package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/gosqs"
	"github.com/Admiral-Piett/goaws/app/router"
)

func main() {
	var filename string
	var debug bool
	var loglevel string
	flag.StringVar(&filename, "config", "", "config file location + name")
	flag.BoolVar(&debug, "debug", false, "set debug log level")
	flag.StringVar(&loglevel, "loglevel", "info", "log level (default info)")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		level, err := log.ParseLevel(loglevel)
		if err != nil {
			log.SetLevel(log.InfoLevel)
			log.Warnf("Failed to parse loglevel %v, defaulting to info", loglevel)
		} else {
			log.SetLevel(level)
		}
	}

	env := "Local"
	if flag.NArg() > 0 {
		env = flag.Arg(0)
	}

	portNumbers := conf.LoadYamlConfig(filename, env)

	if app.CurrentEnvironment.LogToFile {
		filename := app.CurrentEnvironment.LogFile
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Infof("Failed to log to file: %s, using default stderr", filename)
		}
	}

	r := router.New()

	quit := make(chan struct{}, 0)
	go gosqs.PeriodicTasks(1*time.Second, quit)

	utils.InitializeDecoders()

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
