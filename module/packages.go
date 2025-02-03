package module

import (
	"fmt"
	"os/exec"

	"cat2/liftoff/util"
)

func InstallChocoPackages(packages []string, log *util.Logger) error {
	if len(packages) == 0 {
		log.Info("No Chocolatey packages specified")
		return nil
	}

	log.Info("Installing Chocolatey packages...")

	for _, pkg := range packages {
		log.Info(fmt.Sprintf("Installing %s...", pkg))
		cmd := exec.Command("choco", "install", pkg, "-y")

		if output, err := cmd.CombinedOutput(); err != nil {
			log.Error(fmt.Sprintf("Failed to install %s: %s", pkg, string(output)))
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}

		log.Success(fmt.Sprintf("Successfully installed %s", pkg))
	}

	return nil
}
