package main

import (
	"fmt"
	"log"
	"os"

	"github.com/GeneratorEVil/diskmounter/internal/disk"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: diskmounter <image_path> <mount_point>")
		os.Exit(1)
	}

	imagePath := os.Args[1]
	mountPoint := os.Args[2]

	err := disk.Mount(imagePath, mountPoint)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Mounted %s at %s\n", imagePath, mountPoint)
}
