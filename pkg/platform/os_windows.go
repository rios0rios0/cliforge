package platform

import (
	"fmt"
	"os"
	"os/exec"
)

// OSWindows implements OS for Windows systems.
type OSWindows struct{}

func (it *OSWindows) Download(url, tempFilePath string) error {
	return DownloadFile(url, tempFilePath)
}

func (it *OSWindows) Extract(tempFilePath, destPath string) error {
	unzipCmd := exec.Command(
		"powershell",
		"Expand-Archive",
		"-Path",
		tempFilePath,
		"-DestinationPath",
		destPath,
		"-Force",
	)
	unzipCmd.Stderr = os.Stderr
	unzipCmd.Stdout = os.Stdout
	err := unzipCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform decompressing using 'powershell': %w", err)
	}
	return err
}

func (it *OSWindows) Move(tempFilePath, destPath string) error {
	mvCmd := exec.Command("move", tempFilePath, destPath)
	err := mvCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform moving folder using 'move': %w", err)
	}
	return err
}

func (it *OSWindows) Remove(tempFilePath string) error {
	rmCmd := exec.Command("del", tempFilePath)
	err := rmCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform deleting folder using 'del': %w", err)
	}
	return err
}

func (it *OSWindows) MakeExecutable(_ string) error {
	return nil // Windows doesn't need to explicitly make files executable
}

// GetOS returns the platform-specific OS implementation.
func GetOS() *OSWindows {
	return &OSWindows{}
}
