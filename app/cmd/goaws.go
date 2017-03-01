package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/p4tin/goaws/app/conf"
	"github.com/p4tin/goaws/app/router"
)

func main() {
	env := "Local"
	if len(os.Args) == 2 {
		env = os.Args[1]
	}

	var filename string
	flag.StringVar(&filename, "config", "", "config file location + name")
	flag.Parse()

	portNumbers := conf.LoadYamlConfig(filename, env)

	r := router.New()

	if len(portNumbers) == 1 {
		log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[0])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
		log.Fatal(err)
	} else if len(portNumbers) == 2 {
		go func() {
			log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[0])
			err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
			log.Fatal(err)
		}()
		log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[1])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[1], r)
		log.Fatal(err)
	} else {
		log.Fatal("Not enough or too many ports defined to start GoAws.")
	}
}
