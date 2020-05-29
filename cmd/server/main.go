package main


import (
	"flag"
	"github.com/lishengye/sfs/log"
	"github.com/lishengye/sfs/server"
)
import "os"

func main() {
	configFile := flag.String("c", "", "specific config file ")
	if *configFile == "" {
		log.Error("No config file")
		os.Exit(-1)
	}
	flag.Parse()
	// TODO parse configuration from file
	config := server.NewConfig(1234, "/sssss")
	sfsServer := server.NewServer(config)
	log.Info("Starting Server")

	if err := sfsServer.Run(); err != nil {
		log.Error("Server exit error")
	}
}

