package main

import (
	"flag"
	"os"

	"cat2/liftoff/module"
	"cat2/liftoff/util"
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	logger := util.NewLogger(true)

	if *configPath == "" {
		logger.Error("No configuration file specified")
		logger.Info("Usage: liftoff --config <path>")
		os.Exit(1)
	}

	
	if !util.IsAdmin() {
		logger.Error("This program requires administrative privileges")
		os.Exit(1)
	}

	
	config, err := util.LoadConfig(*configPath, logger)
	if err != nil {
		logger.Error("Failed to load configuration")
		os.Exit(1)
	}

	
	if err := util.InstallChocolatey(logger); err != nil {
		if err.Error() != "chocolatey is already installed" {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	
	if err := module.InstallChocoPackages(config.Packages.Chocolatey, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewSystemConfigurator(logger).Configure(config.System); err != nil {
		logger.Error("Failed to apply system configurations")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewEnvironmentManager(logger).Configure(config.Environment); err != nil {
		logger.Error("Failed to configure environment variables")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewWSLManager(logger).Configure(config.WSL); err != nil {
		logger.Error("Failed to configure WSL")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewDownloadManager(logger).Download(config.Downloads); err != nil {
		logger.Error("Failed to download files")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewNetworkManager(logger).Configure(config.Network); err != nil {
		logger.Error("Failed to configure network settings")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if err := module.NewFileManager(logger).ConfigureAssociations(config.FileAssoc); err != nil {
		logger.Error("Failed to configure file associations")
		logger.Error(err.Error())
		os.Exit(1)
	}

	
	if len(config.Git.Repositories) > 0 {
		if err := module.NewGitManager(logger).CloneMultiple(config.Git.Repositories); err != nil {
			logger.Error("Failed to clone repositories")
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	logger.Success("System configuration completed successfully")
}
