package main

import (
	"flag"
	"log"

	"github.com/harshvardha/distributed_file_system/chunkserver"
	"github.com/harshvardha/distributed_file_system/common"
)

func main() {
	port := flag.String("port", "9001", "Port to listen on")
	storage := flag.String("storage", "./storage", "Storage directory path")
	master := flag.String("master", common.MasterAddress, "Master server address")
	flag.Parse()

	address := "localhost:" + *port

	log.Printf("Starting Chunk Server...")
	log.Printf("Address: %s", address)
	log.Printf("Storage: %s", *storage)
	log.Printf("Master: %s", *master)

	server, err := chunkserver.NewServer(address, *storage, *master)
	if err != nil {
		log.Fatalf("Failed to create chunk server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start chunk server: %s", err)
	}
}
