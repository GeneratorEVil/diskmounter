package disk

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Partition represents a disk partition
type Partition struct {
	StartSector int64
	Size        int64
	Type        string
}

// Mount mounts the first partition of the image to the specified mount point
func Mount(imagePath, mountPoint string) error {
	partitions, err := getPartitions(imagePath)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %v", err)
	}
	if len(partitions) == 0 {
		return fmt.Errorf("no partitions found in %s", imagePath)
	}

	// Используем первый раздел
	partition := partitions[0]
	offset := partition.StartSector * 512 // Смещение в байтах

	// Выполняем команду mount
	cmd := exec.Command("mount", "-t", "ext4", "-o", fmt.Sprintf("loop,offset=%d,rw", offset), imagePath, mountPoint)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount failed: %v, %s", err, stderr.String())
	}

	return nil
}

// getPartitions parses the output of fdisk to extract partition info
func getPartitions(imagePath string) ([]Partition, error) {
	cmd := exec.Command("fdisk", "-l", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var partitions []Partition
	re := regexp.MustCompile(`^/.*\d+\s+(\d+)\s+(\d+)\s+(\d+)\s+.*\s+(\w+)$`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			start, _ := strconv.ParseInt(matches[1], 10, 64)
			end, _ := strconv.ParseInt(matches[2], 10, 64)
			sectors, _ := strconv.ParseInt(matches[3], 10, 64)
			partType := matches[4]
			partitions = append(partitions, Partition{
				StartSector: start,
				Size:        sectors * 512, // Размер в байтах
				Type:        partType,
			})
		}
	}

	return partitions, nil
}
