package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// GiteaClient creates a new Gitea client with configuration
func GiteaClient(cfg *Config) (*gitea.Client, error) {
	client, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
	if err != nil {
		ErrorLogger.Printf("Client creation failed for %s", cfg.GiteaURL)
		return nil, fmt.Errorf("create Gitea client: %w", err)
	}
	InfoLogger.Printf("Created client for %s", cfg.GiteaURL)
	return client, nil
}

func parseRepoString(input string) (string, string, error) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format %q - must be owner/repo", input)
	}
	return parts[0], parts[1], nil
}

func WriteYAMLFile(filePath string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	return nil
}

func ReadYAMLFile(filePath string, out interface{}) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}
	return yaml.Unmarshal(b, out)
}

func LoadConfig(filePath string) (*Config, error) {
	if filePath == "" {
		filePath = DefaultConfigFile
	}

	var cfg Config
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := godotenv.Load(); err == nil {
		if token := os.Getenv("GITEA_TOKEN"); token != "" {
			cfg.GiteaToken = token
		}
		if url := os.Getenv("GITEA_URL"); url != "" {
			cfg.GiteaURL = url
		}
	}

	if cfg.GiteaToken == "" {
		return nil, fmt.Errorf("missing Gitea token - configure in file or GITEA_TOKEN env")
	}

	if cfg.GiteaURL == "" {
		return nil, fmt.Errorf("missing Gitea URL - configure in file or GITEA_URL env")
	}

	return &cfg, nil
}
