package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func IsAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

func InstallChocolatey(log *Logger) error {
	if _, err := exec.LookPath("choco"); err == nil {
		return fmt.Errorf("chocolatey is already installed")
	}

	log.Info("Starting Chocolatey installation...")

	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		log.Error("PowerShell not found")
		return fmt.Errorf("powershell not found: %w", err)
	}

	installScript := `Set-ExecutionPolicy Bypass -Scope Process -Force;
	[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
	iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))`
	log.Info("Preparing installation script...")
	cmd := exec.Command(powershell, "-NoProfile", "-InputFormat", "None", "-ExecutionPolicy", "Bypass", "-Command", installScript)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("Failed to get home directory")
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logFile := filepath.Join(homeDir, "chocolatey_install.log")
	f, err := os.Create(logFile)
	if err != nil {
		log.Error("Failed to create log file")
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer f.Close()

	log.Info("Running installation...")

	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Run(); err != nil {
		log.Error("Installation failed")
		return fmt.Errorf("installation failed: %w", err)
	}

	if _, err := exec.LookPath("choco"); err != nil {
		log.Error("Installation verification failed")
		return fmt.Errorf("installation verification failed: %w", err)
	}

	log.Success("Chocolatey has been successfully installed!")

	log.Info("Installing Git...")
	gitCmd := exec.Command("choco", "install", "git", "-y")
	gitCmd.Stdout = f
	gitCmd.Stderr = f

	if err := gitCmd.Run(); err != nil {
		log.Error("Git installation failed")
		return fmt.Errorf("git installation failed: %w", err)
	}

	log.Success("Git has been successfully installed!")
	return nil
}
