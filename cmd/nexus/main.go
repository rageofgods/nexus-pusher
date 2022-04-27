package main

import (
	"fmt"
	"log"
	"net/http"
	"nexus-pusher/pkg/client"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
)

func main() {
	// Load Nexus-Pusher configuration from file
	cfg := config.NewNexusConfig()

	if err := cfg.LoadConfig(configName); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}

	if cfg.Server.Enabled {
		log.Println("Running in server mode.")
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), server.NewRouter()))
	} else {
		log.Println("Running in client mode.")
		client.RunNexusPusher(cfg)
	}

	fmt.Println("test")

	//test, err := client.ReadExport("export.json")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if err := s.UploadComponents(client, test); err != nil {
	//	log.Fatal(err)
	//}

}

const (
	configName string = "config.yaml"
)
