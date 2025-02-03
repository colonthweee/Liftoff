package module

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cat2/liftoff/types"
	"cat2/liftoff/util"
)

type GitManager struct {
	log *util.Logger
}

func NewGitManager(log *util.Logger) *GitManager {
	return &GitManager{
		log: log,
	}
}

func (g *GitManager) CloneMultiple(repositories []types.Repository) error {
	var errors []string

	
	failedDir := filepath.Join(os.ExpandEnv("${USERPROFILE}"), "Liftoff", "Failed")
	if err := os.MkdirAll(failedDir, 0755); err != nil {
		return fmt.Errorf("failed to create failed operations directory: %w", err)
	}

	for _, repo := range repositories {
		if err := g.Clone(repo); err != nil {
			errors = append(errors, fmt.Sprintf("failed to clone %s: %v", repo.URL, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some repositories failed to clone:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

func (g *GitManager) Clone(config types.Repository) error {
	expandedPath := os.ExpandEnv(config.Path)

	
	if files, err := os.ReadDir(expandedPath); err == nil && len(files) > 0 {
		
		timestamp := time.Now().Format("20060102_150405")
		repoName := filepath.Base(expandedPath)
		failedDir := filepath.Join(os.ExpandEnv("${USERPROFILE}"), "Liftoff", "Failed")
		newPath := filepath.Join(failedDir, fmt.Sprintf("%s_%s", repoName, timestamp))

		
		if err := os.MkdirAll(failedDir, 0755); err != nil {
			return fmt.Errorf("failed to create failed directory: %w", err)
		}

		
		expandedPath = newPath
		g.log.Warn(fmt.Sprintf("Original directory not empty, redirecting clone to: %s", expandedPath))

		
		metadataPath := filepath.Join(failedDir, fmt.Sprintf("%s_%s.txt", repoName, timestamp))
		metadata := fmt.Sprintf("Original Path: %s\nRedirected Path: %s\nURL: %s\nBranch: %s\nDepth: %d\nReason: Original directory not empty\n",
			config.Path, expandedPath, config.URL, config.Branch, config.Depth)

		if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
			g.log.Warn(fmt.Sprintf("Failed to write metadata: %v", err))
		}
	}

	
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("invalid git URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	host := strings.ToLower(parsedURL.Host)
	trustedHosts := map[string]bool{
		"github.com":    true,
		"gitlab.com":    true,
		"bitbucket.org": true,
		"dev.azure.com": true,
	}

	if !trustedHosts[host] {
		return fmt.Errorf("untrusted Git host: %s", host)
	}

	parentDir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	args := []string{"clone"}

	
	args = append(args,
		"--config", "protocol.version=2",
		"--config", "transfer.fsckObjects=true",
		"--config", "fetch.fsckObjects=true",
	)

	if config.Branch != "" {
		args = append(args, "-b", config.Branch)
	}
	if config.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", config.Depth))
	}

	args = append(args, config.URL, expandedPath)

	g.log.Info(fmt.Sprintf("Cloning %s into %s", config.URL, expandedPath))

	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSL_NO_VERIFY=false",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s", string(output))
	}

	if config.SubmoduleInit {
		g.log.Info("Initializing submodules")
		cmd = exec.Command("git", "-C", expandedPath, "submodule", "update",
			"--init", "--recursive",
			"--config", "protocol.version=2",
			"--config", "transfer.fsckObjects=true",
			"--config", "fetch.fsckObjects=true")

		cmd.Env = append(os.Environ(),
			"GIT_TERMINAL_PROMPT=0",
			"GIT_SSL_NO_VERIFY=false",
		)

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("submodule initialization failed: %s", string(output))
		}
	}

	g.log.Success(fmt.Sprintf("Successfully cloned repository to %s", expandedPath))
	return nil
}
