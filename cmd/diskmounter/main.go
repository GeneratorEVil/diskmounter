package main

import (
	"fmt"
	"log"
	"os"

	"github.com/GeneratorEVil/diskmounter/internal/disk"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: diskmounter <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  mount <image_path> <mount_point>")
		fmt.Println("  umount <mount_point>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "mount":
		if len(os.Args) != 4 {
			fmt.Println("Usage: diskmounter mount <image_path> <mount_point>")
			os.Exit(1)
		}
		imagePath := os.Args[2]
		mountPoint := os.Args[3]
		err := disk.Mount(imagePath, mountPoint)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("Mounted %s at %s\n", imagePath, mountPoint)

	case "umount":
		if len(os.Args) != 3 {
			fmt.Println("Usage: diskmounter umount <mount_point>")
			os.Exit(1)
		}
		mountPoint := os.Args[2]
		err := disk.Unmount(mountPoint)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("Unmounted %s\n", mountPoint)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
