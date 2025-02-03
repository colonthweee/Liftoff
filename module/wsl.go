package module

import (
	"fmt"
	"os/exec"
	"strings"

	"cat2/liftoff/types"
	"cat2/liftoff/util"
)

type WSLManager struct {
	log *util.Logger
}

func NewWSLManager(log *util.Logger) *WSLManager {
	return &WSLManager{
		log: log,
	}
}

func (w *WSLManager) Configure(config types.WSLConfig) error {
	
	if !w.isWSLAvailable() {
		w.log.Error("WSL is not available. Please enable it first using Windows Features")
		return fmt.Errorf("WSL is not available")
	}

	
	for _, dist := range config.Distributions {
		if err := w.installDistribution(dist); err != nil {
			return err
		}
	}

	
	if config.DefaultDistro != "" {
		if err := w.setDefaultDistribution(config.DefaultDistro); err != nil {
			return err
		}
	}

	return nil
}

func (w *WSLManager) isWSLAvailable() bool {
	cmd := exec.Command("wsl", "--status")
	return cmd.Run() == nil
}

func (w *WSLManager) isDistributionInstalled(name string) bool {
	cmd := exec.Command("wsl", "-l", "-q")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	installedDistros := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, distro := range installedDistros {
		if strings.TrimSpace(distro) == name {
			return true
		}
	}
	return false
}

func (w *WSLManager) installDistribution(dist types.WSLDistribution) error {
	if w.isDistributionInstalled(dist.Name) {
		w.log.Info(fmt.Sprintf("Distribution %s is already installed", dist.Name))
		return nil
	}

	w.log.Info(fmt.Sprintf("Installing WSL distribution: %s", dist.Name))

	var cmd *exec.Cmd
	if dist.Version == "latest" || dist.Version == "" {
		cmd = exec.Command("wsl", "--install", "-d", dist.Name)
	} else {
		
		
		w.log.Warn(fmt.Sprintf("Version specification is not supported. Installing latest version of %s", dist.Name))
		cmd = exec.Command("wsl", "--install", "-d", dist.Name)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		w.log.Error(fmt.Sprintf("Failed to install %s: %s", dist.Name, string(output)))
		return fmt.Errorf("failed to install %s: %w", dist.Name, err)
	}

	w.log.Success(fmt.Sprintf("Successfully installed %s", dist.Name))
	return nil
}

func (w *WSLManager) setDefaultDistribution(name string) error {
	if !w.isDistributionInstalled(name) {
		w.log.Error(fmt.Sprintf("Distribution %s is not installed", name))
		return fmt.Errorf("distribution %s is not installed", name)
	}

	w.log.Info(fmt.Sprintf("Setting %s as default WSL distribution", name))

	cmd := exec.Command("wsl", "--set-default", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		w.log.Error(fmt.Sprintf("Failed to set default distribution: %s", string(output)))
		return fmt.Errorf("failed to set default distribution: %w", err)
	}

	w.log.Success(fmt.Sprintf("Successfully set %s as default distribution", name))
	return nil
}
