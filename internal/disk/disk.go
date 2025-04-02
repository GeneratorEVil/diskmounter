package disk

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
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
	// Проверяем, является ли файл VDI
	if isVDI(imagePath) {
		return mountVDI(imagePath, mountPoint)
	}

	// Обычная логика для RAW-образов
	partitions, err := getPartitions(imagePath)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %v", err)
	}
	if len(partitions) == 0 {
		return fmt.Errorf("no partitions found in %s", imagePath)
	}

	partition := partitions[0]
	offset := partition.StartSector * 512

	cmd := exec.Command("mount", "-t", "ext4", "-o", fmt.Sprintf("loop,offset=%d,rw", offset), imagePath, mountPoint)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount failed: %v, %s", err, stderr.String())
	}
	return nil
}

// Unmount unmounts the specified mount point
func Unmount(mountPoint string) error {
	cmd := exec.Command("umount", mountPoint)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("umount failed: %v, %s", err, stderr.String())
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
			_, _ = strconv.ParseInt(matches[2], 10, 64)
			sectors, _ := strconv.ParseInt(matches[3], 10, 64)
			partType := matches[4]
			partitions = append(partitions, Partition{
				StartSector: start,
				Size:        sectors * 512,
				Type:        partType,
			})
		}
	}

	return partitions, nil
}

// isVDI checks if the file is a VDI image based on extension
func isVDI(imagePath string) bool {
	return strings.ToLower(filepath.Ext(imagePath)) == ".vdi"
}

// mountVDI mounts a VDI image using qemu-nbd
func mountVDI(imagePath, mountPoint string) error {
	// Загружаем модуль nbd, если не загружен
	if err := exec.Command("modprobe", "nbd").Run(); err != nil {
		return fmt.Errorf("failed to load nbd module: %v", err)
	}

	// Находим свободное nbd-устройство
	nbdDevice := "/dev/nbd0" // Можно сделать динамический поиск
	cmd := exec.Command("qemu-nbd", "-c", nbdDevice, imagePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to connect VDI to %s: %v, %s", nbdDevice, err, stderr.String())
	}

	// Считываем разделы с nbd-устройства
	partitions, err := getPartitions(nbdDevice)
	if err != nil {
		_ = exec.Command("qemu-nbd", "-d", nbdDevice).Run() // Отключаем в случае ошибки
		return fmt.Errorf("failed to get partitions from %s: %v", nbdDevice, err)
	}
	if len(partitions) == 0 {
		_ = exec.Command("qemu-nbd", "-d", nbdDevice).Run()
		return fmt.Errorf("no partitions found in %s", imagePath)
	}

	// Монтируем первый раздел
	partitionDevice := nbdDevice + "p1" // Например, /dev/nbd0p1
	cmd = exec.Command("mount", "-t", "ext4", partitionDevice, mountPoint)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = exec.Command("qemu-nbd", "-d", nbdDevice).Run()
		return fmt.Errorf("mount failed: %v, %s", err, stderr.String())
	}

	// Оставляем подключение активным, демонтирование обработает отключение
	return nil
}
