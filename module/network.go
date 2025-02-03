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

type NetworkManager struct {
	log *util.Logger
}

func NewNetworkManager(log *util.Logger) *NetworkManager {
	return &NetworkManager{
		log: log,
	}
}

func (n *NetworkManager) Configure(config types.NetworkConfig) error {
	
	if len(config.DNSServers) > 0 {
		if err := n.setDNSServers(config.DNSServers); err != nil {
			return err
		}
	}

	
	if len(config.HostsEntries) > 0 {
		if err := n.updateHostsFile(config.HostsEntries); err != nil {
			return err
		}
	}

	
	if config.Proxy.Enable {
		if err := n.setProxy(config.Proxy); err != nil {
			return err
		}
	}

	return nil
}

func (n *NetworkManager) setDNSServers(servers []string) error {
	n.log.Info("Setting DNS servers")

	
	cmd := exec.Command("netsh", "interface", "show", "interface")
	output, err := cmd.Output()
	if err != nil {
		n.log.Error("Failed to get network interfaces")
		return fmt.Errorf("failed to get network interfaces: %w", err)
	}

	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Enabled") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				interfaceName := strings.Join(fields[3:], " ")
				
				dnsCmd := exec.Command("netsh", "interface", "ipv4", "set", "dns",
					interfaceName, "static", servers[0])
				if err := dnsCmd.Run(); err != nil {
					n.log.Error(fmt.Sprintf("Failed to set primary DNS for %s", interfaceName))
					return fmt.Errorf("failed to set DNS: %w", err)
				}

				
				for i, server := range servers[1:] {
					addCmd := exec.Command("netsh", "interface", "ipv4", "add", "dns",
						interfaceName, server, fmt.Sprintf("index=%d", i+2))
					if err := addCmd.Run(); err != nil {
						n.log.Error(fmt.Sprintf("Failed to add DNS server %s", server))
						return fmt.Errorf("failed to add DNS server: %w", err)
					}
				}
			}
		}
	}

	n.log.Success("Successfully configured DNS servers")
	return nil
}

func (n *NetworkManager) updateHostsFile(entries map[string]string) error {
	n.log.Info("Updating hosts file")

	hostsPath := `C:\Windows\System32\drivers\etc\hosts`

	
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		n.log.Error("Failed to read hosts file")
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	
	lines := strings.Split(string(content), "\n")
	existingEntries := make(map[string]bool)
	newLines := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				existingEntries[fields[1]] = true
			}
		}
		newLines = append(newLines, line)
	}

	
	for hostname, ip := range entries {
		if !existingEntries[hostname] {
			newLines = append(newLines, fmt.Sprintf("%s\t%s", ip, hostname))
		}
	}

	
	if err := os.WriteFile(hostsPath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		n.log.Error("Failed to write hosts file")
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	n.log.Success("Successfully updated hosts file")
	return nil
}

func (n *NetworkManager) setProxy(config types.ProxyConfig) error {
	n.log.Info("Configuring proxy settings")

	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.ALL_ACCESS)
	if err != nil {
		n.log.Error("Failed to open registry key")
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	
	if err := key.SetDWordValue("ProxyEnable", uint32(1)); err != nil {
		n.log.Error("Failed to enable proxy")
		return fmt.Errorf("failed to enable proxy: %w", err)
	}

	
	proxyServer := fmt.Sprintf("%s:%d", config.Server, config.Port)
	if err := key.SetStringValue("ProxyServer", proxyServer); err != nil {
		n.log.Error("Failed to set proxy server")
		return fmt.Errorf("failed to set proxy server: %w", err)
	}

	n.log.Success("Successfully configured proxy settings")
	return nil
}
