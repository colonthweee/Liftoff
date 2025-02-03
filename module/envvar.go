package module

import (
	"fmt"
	"os"
	"strings"

	"cat2/liftoff/types"
	"cat2/liftoff/util"

	"golang.org/x/sys/windows/registry"
)

type EnvironmentManager struct {
	log *util.Logger
}

func NewEnvironmentManager(log *util.Logger) *EnvironmentManager {
	return &EnvironmentManager{
		log: log,
	}
}

func (e *EnvironmentManager) Configure(config types.EnvironmentConfig) error {
	
	if len(config.PathAppend) > 0 {
		if err := e.appendToPath(config.PathAppend); err != nil {
			return err
		}
	}

	
	if len(config.Variables) > 0 {
		if err := e.setVariables(config.Variables); err != nil {
			return err
		}
	}

	return nil
}

func (e *EnvironmentManager) appendToPath(paths []string) error {
	e.log.Info("Configuring PATH variable")

	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.ALL_ACCESS)
	if err != nil {
		e.log.Error("Failed to open Environment registry key")
		return fmt.Errorf("failed to open Environment registry key: %w", err)
	}
	defer key.Close()

	
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		e.log.Error("Failed to read PATH variable")
		return fmt.Errorf("failed to read PATH variable: %w", err)
	}

	
	var pathComponents []string
	if currentPath != "" {
		pathComponents = strings.Split(currentPath, ";")
	}

	
	pathMap := make(map[string]bool)
	for _, p := range pathComponents {
		pathMap[strings.TrimSpace(p)] = true
	}

	modified := false
	for _, newPath := range paths {
		expandedPath := os.ExpandEnv(newPath)
		if !pathMap[expandedPath] {
			pathComponents = append(pathComponents, expandedPath)
			pathMap[expandedPath] = true
			modified = true
			e.log.Success(fmt.Sprintf("Added %s to PATH", expandedPath))
		}
	}

	
	if modified {
		newPath := strings.Join(pathComponents, ";")
		if err := key.SetStringValue("Path", newPath); err != nil {
			e.log.Error("Failed to update PATH variable")
			return fmt.Errorf("failed to update PATH variable: %w", err)
		}
	}

	return nil
}

func (e *EnvironmentManager) setVariables(variables map[string]string) error {
	e.log.Info("Setting environment variables")

	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.ALL_ACCESS)
	if err != nil {
		e.log.Error("Failed to open Environment registry key")
		return fmt.Errorf("failed to open Environment registry key: %w", err)
	}
	defer key.Close()

	for name, value := range variables {
		expandedValue := os.ExpandEnv(value)
		if err := key.SetStringValue(name, expandedValue); err != nil {
			e.log.Error(fmt.Sprintf("Failed to set %s", name))
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
		e.log.Success(fmt.Sprintf("Set %s=%s", name, expandedValue))
	}

	return nil
}
