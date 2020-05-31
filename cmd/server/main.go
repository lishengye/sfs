package main

import (
	"flag"
	"github.com/lishengye/sfs/log"
	"github.com/lishengye/sfs/server"
)
import "os"

func main() {
	configFile := flag.String("c", "", "specific config file ")
	flag.Parse()
	if *configFile == "" {
		log.Error("No config file")
		os.Exit(-1)
	}

	config, err := server.NewConfig(*configFile)
	if err != nil {
		log.Error("Read ConfigFile error: %s", err.Error())
		return
	}
	sfsServer := server.NewServer(config)
	log.Info("Starting Server")

	if err := sfsServer.Run(); err != nil {
		log.Error("Server exit error")
	}
}
