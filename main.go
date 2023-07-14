package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mrkhay/gobank/api"
	"github.com/mrkhay/gobank/storage"
)

func main() {

	port := flag.String("p", "", "specify port number")
	flag.Parse()

	store, err := storage.NewPostgresStorage()
	if err != nil {
		log.Fatal("Failed to connect - ", err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *port == "" {
		log.Fatal("port address required")
	}

	// instace of server
	server := api.NewApiServer(fmt.Sprintf(":%s", *port), store)
	server.Run()

}
