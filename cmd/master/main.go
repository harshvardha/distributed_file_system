package main

import (
	"log"

	"github.com/harshvardha/distributed_file_system/common"
	"github.com/harshvardha/distributed_file_system/master"
)

func main() {
	log.Println("Starting Distributed File System Master Server...")

	server := master.NewServer(common.MasterAddress)
	if err := server.Start(); err != nil {
		log.Fatalf("Master server failed: %v", err)
	}
}
