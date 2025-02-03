package module

import (
	"cat2/liftoff/types"
	"cat2/liftoff/util"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type FileManager struct {
	log *util.Logger
}

func NewFileManager(log *util.Logger) *FileManager {
	return &FileManager{
		log: log,
	}
}

func (f *FileManager) ConfigureAssociations(config types.FileAssocConfig) error {
	for ext, program := range config.Associations {
		if err := f.setFileAssociation(ext, program); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileManager) setFileAssociation(ext, program string) error {
	
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	
	program = os.ExpandEnv(program)

	
	if _, err := os.Stat(program); os.IsNotExist(err) {
		f.log.Error(fmt.Sprintf("Program not found: %s", program))
		return fmt.Errorf("program not found: %s", program)
	}

	f.log.Info(fmt.Sprintf("Setting file association for %s to %s", ext, program))

	
	progID := fmt.Sprintf("liftoff%s", strings.Replace(ext, ".", "", 1))

	
	extKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, ext, registry.ALL_ACCESS)
	if err != nil {
		f.log.Error(fmt.Sprintf("Failed to create registry key for %s", ext))
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer extKey.Close()

	if err := extKey.SetStringValue("", progID); err != nil {
		f.log.Error("Failed to set default value for extension")
		return fmt.Errorf("failed to set default value: %w", err)
	}

	
	progIDKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, progID, registry.ALL_ACCESS)
	if err != nil {
		f.log.Error("Failed to create ProgID key")
		return fmt.Errorf("failed to create ProgID key: %w", err)
	}
	defer progIDKey.Close()

	
	description := fmt.Sprintf("%s File", strings.ToUpper(ext[1:]))
	if err := progIDKey.SetStringValue("", description); err != nil {
		f.log.Error("Failed to set ProgID description")
		return fmt.Errorf("failed to set ProgID description: %w", err)
	}

	
	shellKey, _, err := registry.CreateKey(progIDKey, "shell\\open\\command", registry.ALL_ACCESS)
	if err != nil {
		f.log.Error("Failed to create shell command key")
		return fmt.Errorf("failed to create shell command key: %w", err)
	}
	defer shellKey.Close()

	
	command := fmt.Sprintf("\"%s\" \"%%1\"", program)
	if err := shellKey.SetStringValue("", command); err != nil {
		f.log.Error("Failed to set shell command")
		return fmt.Errorf("failed to set shell command: %w", err)
	}

	
	shellChangeNotify()

	f.log.Success(fmt.Sprintf("Successfully associated %s with %s", ext, program))
	return nil
}


func shellChangeNotify() {
	
	exec.Command("cmd", "/c", "assoc", "/c").Run()
}
