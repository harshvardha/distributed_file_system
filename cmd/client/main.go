package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/harshvardha/distributed_file_system/client"
	"github.com/harshvardha/distributed_file_system/common"
)

func main() {
	// Creating subcommands
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadFile := uploadCmd.String("file", "", "Local file path to upload")
	uploadName := uploadCmd.String("name", "", "Remote file name")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downloadName := downloadCmd.String("name", "", "Remote file name to download")
	downloadOutput := downloadCmd.String("output", "", "Local output file path")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Creating client
	dfsClient := client.NewClient(common.MasterAddress)

	// Parsing subcommands
	switch os.Args[1] {
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		if *uploadFile == "" || *uploadName == "" {
			uploadCmd.PrintDefaults()
			os.Exit(1)
		}

		if err := dfsClient.UploadFile(*uploadFile, *uploadName); err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
		fmt.Printf("Successfully uploaded: %s\n", *uploadName)
	case "download":
		downloadCmd.Parse(os.Args[2:])
		if *downloadName == "" || *downloadOutput == "" {
			downloadCmd.PrintDefaults()
			os.Exit(1)
		}

		if err := dfsClient.DownloadFile(*downloadName, *downloadOutput); err != nil {
			log.Fatalf("Download failed: %v", err)
		}
		fmt.Printf("Successfully downloaded to: %s\n", *downloadOutput)
	case "list":
		listCmd.Parse(os.Args[2:])

		files, err := dfsClient.ListFiles()
		if err != nil {
			log.Fatalf("List failed: %v", err)
		}

		if len(files) == 0 {
			fmt.Println("No files in the system")
		} else {
			fmt.Printf("Files in DFS (%d total):\n", len(files))
			fmt.Println("----------------------------------------")
			for _, file := range files {
				fmt.Printf("Name: %s\n", file.Filename)
				fmt.Printf("Size: %d bytes\n", file.Filesize)
				fmt.Printf("Chunks: %d\n", file.NumChunks)
				fmt.Println("----------------------------------------")
			}
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Distributed File System Client")
	fmt.Println("\nUsage:")
	fmt.Println("	client upload -file <local_path> -name <remote_name>")
	fmt.Println("	client download -name <remote_name> -output <local_path>")
	fmt.Println("	client list")
	fmt.Println("\nExamples:")
	fmt.Println("	client upload -file ./test.txt -name myfile.txt")
	fmt.Println("	client download -name myfile.txt -output ./downloaded.txt")
	fmt.Println("	client list")
}
