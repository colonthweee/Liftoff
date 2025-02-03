package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cat2/liftoff/types"
	"cat2/liftoff/util"

	"golang.org/x/sys/windows/registry"
)


type SystemConfigurator struct {
	log *util.Logger
}


func NewSystemConfigurator(log *util.Logger) *SystemConfigurator {
	return &SystemConfigurator{
		log: log,
	}
}


func (s *SystemConfigurator) CreateFolders(paths []string) error {
	for _, path := range paths {
		expanded := os.ExpandEnv(path)
		s.log.Info(fmt.Sprintf("Creating folder: %s", expanded))

		if err := os.MkdirAll(expanded, 0755); err != nil {
			s.log.Error(fmt.Sprintf("Failed to create folder: %s", expanded))
			return fmt.Errorf("failed to create folder %s: %w", expanded, err)
		}

		s.log.Success(fmt.Sprintf("Created folder: %s", expanded))
	}
	return nil
}


func (s *SystemConfigurator) CreateFiles(files map[string]string) error {
	for path, content := range files {
		expanded := os.ExpandEnv(path)
		s.log.Info(fmt.Sprintf("Creating file: %s", expanded))

		
		dir := filepath.Dir(expanded)
		if err := os.MkdirAll(dir, 0755); err != nil {
			s.log.Error(fmt.Sprintf("Failed to create directory for file: %s", expanded))
			return fmt.Errorf("failed to create directory for %s: %w", expanded, err)
		}

		if err := os.WriteFile(expanded, []byte(content), 0644); err != nil {
			s.log.Error(fmt.Sprintf("Failed to create file: %s", expanded))
			return fmt.Errorf("failed to create file %s: %w", expanded, err)
		}

		s.log.Success(fmt.Sprintf("Created file: %s", expanded))
	}
	return nil
}


func (s *SystemConfigurator) SetRegistryValue(config types.RegistryConfig) error {
	s.log.Info(fmt.Sprintf("Setting registry value: %s\\%s", config.Path, config.Name))

	
	var root registry.Key
	switch strings.ToUpper(config.Root) {
	case "HKEY_LOCAL_MACHINE", "HKLM":
		root = registry.LOCAL_MACHINE
	case "HKEY_CURRENT_USER", "HKCU":
		root = registry.CURRENT_USER
	case "HKEY_USERS", "HKU":
		root = registry.USERS
	case "HKEY_CLASSES_ROOT", "HKCR":
		root = registry.CLASSES_ROOT
	default:
		return fmt.Errorf("invalid registry root: %s", config.Root)
	}

	
	key, err := registry.OpenKey(root, config.Path, registry.ALL_ACCESS)
	if err != nil {
		
		key, _, err = registry.CreateKey(root, config.Path, registry.ALL_ACCESS)
		if err != nil {
			s.log.Error(fmt.Sprintf("Failed to open/create registry key: %s", config.Path))
			return fmt.Errorf("failed to access registry key: %w", err)
		}
	}
	defer key.Close()

	
	switch strings.ToLower(config.Type) {
	case "string", "sz":
		err = key.SetStringValue(config.Name, config.Value.(string))
	case "dword":
		
		val, ok := config.Value.(int64)
		if !ok {
			return fmt.Errorf("invalid DWORD value for %s", config.Name)
		}
		err = key.SetDWordValue(config.Name, uint32(val))
	case "binary":
		
		val, ok := config.Value.([]byte)
		if !ok {
			return fmt.Errorf("invalid binary value for %s", config.Name)
		}
		err = key.SetBinaryValue(config.Name, val)
	default:
		return fmt.Errorf("unsupported registry value type: %s", config.Type)
	}

	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to set registry value: %s", config.Name))
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	s.log.Success(fmt.Sprintf("Set registry value: %s\\%s", config.Path, config.Name))
	return nil
}


func (s *SystemConfigurator) SetDarkMode(enable bool) error {
	
	const personalizePath = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`

	darkModeConfig := types.RegistryConfig{
		Root:  "HKCU",
		Path:  personalizePath,
		Name:  "AppsUseLightTheme",
		Type:  "dword",
		Value: int64(0),
	}

	if !enable {
		darkModeConfig.Value = int64(1)
	}

	s.log.Info(fmt.Sprintf("Setting dark mode to: %v", enable))
	return s.SetRegistryValue(darkModeConfig)
}


func (s *SystemConfigurator) Configure(config types.SystemConfig) error {
	
	if len(config.Folders) > 0 {
		if err := s.CreateFolders(config.Folders); err != nil {
			return err
		}
	}

	
	if len(config.Files) > 0 {
		if err := s.CreateFiles(config.Files); err != nil {
			return err
		}
	}

	
	for _, reg := range config.Registry {
		if err := s.SetRegistryValue(reg); err != nil {
			return err
		}
	}

	
	if config.DarkMode {
		if err := s.SetDarkMode(true); err != nil {
			return err
		}
	}

	return nil
}
