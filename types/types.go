package types


type Config struct {
	Packages    PackageConfig     `toml:"packages"`
	System      SystemConfig      `toml:"system"`
	Git         GitConfig         `toml:"git"`
	Environment EnvironmentConfig `toml:"environment"`
	WSL         WSLConfig         `toml:"wsl"`
	Downloads   DownloadConfig    `toml:"downloads"`
	Network     NetworkConfig     `toml:"network"`
	FileAssoc   FileAssocConfig   `toml:"file_associations"`
}


type GitConfig struct {
	Repositories []Repository `toml:"repositories"`
}


type Repository struct {
	URL           string `toml:"url"`
	Path          string `toml:"path"`
	Branch        string `toml:"branch,omitempty"`
	Depth         int    `toml:"depth,omitempty"`
	SubmoduleInit bool   `toml:"submodule_init,omitempty"`
}


type PackageConfig struct {
	Chocolatey []string `toml:"chocolatey"`
	Winget     []string `toml:"winget"`
}


type SystemConfig struct {
	DarkMode bool              `toml:"dark_mode"`
	Folders  []string          `toml:"folders"`
	Files    map[string]string `toml:"files"`
	Registry []RegistryConfig  `toml:"registry"`
}


type RegistryConfig struct {
	Root  string      `toml:"root"`  
	Path  string      `toml:"path"`  
	Name  string      `toml:"name"`  
	Type  string      `toml:"type"`  
	Value interface{} `toml:"value"` 
}

type EnvironmentConfig struct {
	PathAppend []string          `toml:"path_append"`
	Variables  map[string]string `toml:"variables"`
}


type WSLConfig struct {
	DefaultDistro string            `toml:"default_distro"`
	Distributions []WSLDistribution `toml:"distributions"`
}


type WSLDistribution struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type DownloadConfig struct {
	Files []DownloadFile `toml:"files"`
}

type DownloadFile struct {
	URL    string `toml:"url"`
	Dest   string `toml:"dest"`
	SHA256 string `toml:"sha256,omitempty"`
	Rename string `toml:"rename,omitempty"`
}


type NetworkConfig struct {
	DNSServers   []string          `toml:"dns_servers"`
	HostsEntries map[string]string `toml:"hosts_entries"`
	Proxy        ProxyConfig       `toml:"proxy,omitempty"`
}

type ProxyConfig struct {
	Enable   bool   `toml:"enable"`
	Server   string `toml:"server"`
	Port     int    `toml:"port"`
	Username string `toml:"username,omitempty"`
	Password string `toml:"password,omitempty"`
}


type FileAssocConfig struct {
	Associations map[string]string `toml:"associations"` 
}
