package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// InitZDriveMount orchestrates the automated mounting of the Universal Z-Drive
// via the Hetzner Storage Box using Rclone over SFTP.
func InitZDriveMount() error {
	fmt.Println("[Z-DRIVE] Initializing Universal Mesh Storage Mount...")

	// 1. Verify/Install Rclone (mocking the download process for this deployment)
	// In production, this would `Invoke-WebRequest` the rclone.zip if not found.
	rclonePath := checkRclone()
	if rclonePath == "" {
		fmt.Println("[Z-DRIVE] WARNING: rclone executable not found in PATH or local directory.")
		fmt.Println("[Z-DRIVE] Please ensure rclone is installed to enable universal Z:\\ mounting.")
		return fmt.Errorf("rclone not found")
	}

	// 2. Configure the SFTP remote
	configDir := filepath.Join(os.Getenv("USERPROFILE"), ".config", "rclone")
	if runtime.GOOS != "windows" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "rclone")
	}
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "rclone.conf")
	
	// This hardcodes the config using the provided credentials
	// Note: Password normally obfuscated via `rclone obscure` but for standard programmatic setup we can rely on standard rclone configs.
	configContent := `
[hetzner_zdrive]
type = sftp
host = u589955.your-storagebox.de
user = u589955
port = 23
pass = uY1l4xW73UovP6Jv6c6gGv8jU6e7I2QyF2w8e4_R-uW0p9iO2Fw5k6
shell_type = unix
md5sum_command = md5sum
sha1sum_command = sha1sum
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write rclone config: %v", err)
	}

	// 3. Mount the drive in the background
	mountPoint := "Z:"
	if runtime.GOOS != "windows" {
		mountPoint = "/mnt/zdrive"
		os.MkdirAll(mountPoint, 0755)
	}

	fmt.Printf("[Z-DRIVE] Spawning rclone mount process to %s\n", mountPoint)
	
	// Command: rclone mount hetzner_zdrive: /mnt/zdrive --vfs-cache-mode writes
	cmd := exec.Command(rclonePath, "mount", "hetzner_zdrive:", mountPoint, "--vfs-cache-mode", "writes", "--daemon")
	if runtime.GOOS == "windows" {
		// Windows rclone mount doesn't support --daemon natively, requires WinFsp
		cmd = exec.Command(rclonePath, "mount", "hetzner_zdrive:", mountPoint, "--vfs-cache-mode", "writes")
		// Start async
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start Windows Z: drive mount: %v", err)
		}
	} else {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to daemonize mount: %v", err)
		}
	}

	fmt.Printf("[Z-DRIVE] Successfully mounted Hetzner Storage Box to %s\n", mountPoint)
	return nil
}

func checkRclone() string {
	path, err := exec.LookPath("rclone")
	if err == nil {
		return path
	}
	// Check current directory
	if _, err := os.Stat("rclone.exe"); err == nil {
		return "rclone.exe"
	}
	return ""
}
