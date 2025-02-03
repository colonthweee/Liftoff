package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cat2/liftoff/types"

	"gopkg.in/yaml.v3"
)


func LoadConfig(path string, log *Logger) (*types.Config, error) {
	log.Info("Loading configuration from " + path)

	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("Configuration file not found")
		return nil, err
	}

	
	data, err := os.ReadFile(path)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to read configuration file: %v", err))
		return nil, err
	}

	
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Error(fmt.Sprintf("Failed to parse configuration file: %v", err))
		
		log.Debug("YAML content:")
		log.Debug(string(data))
		return nil, err
	}

	
	if err := validateConfig(&config, log); err != nil {
		log.Error(fmt.Sprintf("Failed to validate configuration: %v", err))
		return nil, err
	}

	log.Success("Configuration loaded successfully")
	return &config, nil
}


func validateConfig(config *types.Config, log *Logger) error {
	
	if len(config.System.Folders) > 0 {
		for i, folder := range config.System.Folders {
			config.System.Folders[i] = os.ExpandEnv(folder)
		}
	}

	
	for path, content := range config.System.Files {
		expandedPath := os.ExpandEnv(path)
		delete(config.System.Files, path)
		config.System.Files[expandedPath] = content
	}

	
	for key, value := range config.Environment.Variables {
		config.Environment.Variables[key] = os.ExpandEnv(value)
	}
	for i, path := range config.Environment.PathAppend {
		config.Environment.PathAppend[i] = os.ExpandEnv(path)
	}

	
	for i, file := range config.Downloads.Files {
		config.Downloads.Files[i].Dest = os.ExpandEnv(file.Dest)
	}

	
	expandedAssociations := make(map[string]string)
	for ext, program := range config.FileAssoc.Associations {
		
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		
		expandedPath := os.ExpandEnv(program)
		
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			log.Warn(fmt.Sprintf("Program not found for %s: %s", ext, expandedPath))
		}
		expandedAssociations[ext] = expandedPath
	}
	config.FileAssoc.Associations = expandedAssociations

	
	for i, repo := range config.Git.Repositories {
		config.Git.Repositories[i].Path = os.ExpandEnv(repo.Path)
	}

	
	if len(config.Packages.Chocolatey) == 0 &&
		len(config.System.Folders) == 0 &&
		len(config.Git.Repositories) == 0 {
		log.Warn("Configuration appears to be empty or missing key sections")
	}

	return nil
}


func expandAndAbsPath(path string) (string, error) {
	expanded := os.ExpandEnv(path)
	if !filepath.IsAbs(expanded) {
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return expanded, nil
}
