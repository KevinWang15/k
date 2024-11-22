package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KevinWang15/k/pkg/model"
	"github.com/KevinWang15/k/pkg/utils"
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
)

var ImportCommand = &cobra.Command{
	Use:   "import",
	Short: "Import all existing configs from KUBECONFIG",
	Run: func(cmd *cobra.Command, args []string) {
		if err := importKubeconfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error importing kubeconfig: %v\n", err)
			os.Exit(1)
		}
	},
}

func readAndEncodeFile(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	// Handle both absolute paths and paths relative to kubeconfig
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(getKubeConfigPath()), path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return data, nil
}

func importKubeconfig() error {
	// Load existing kubeconfig - this handles both YAML and JSON formats
	kubeconfig, err := clientcmd.LoadFromFile(getKubeConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	configPath := utils.GetConfigPath()

	// Load existing config or create new one
	config := &model.Config{
		Shortcuts: make(map[string]string),
		Clusters:  []model.Cluster{},
	}

	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if len(configData) == 0 {
			configData = []byte("{}")
		}
		if err := json.Unmarshal(configData, config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Convert kubeconfig clusters to our model
	for name, cluster := range kubeconfig.Clusters {
		newCluster := model.Cluster{
			Name:    name,
			Cluster: model.FromAPICluster(cluster),
		}

		for _, context := range kubeconfig.Contexts {
			if context.Cluster == name {
				if authInfo, exists := kubeconfig.AuthInfos[context.AuthInfo]; exists {
					newCluster.User = model.FromAPIAuthInfo(authInfo)
				}
			}
		}

		// Check if cluster already exists
		exists := false
		for i, existing := range config.Clusters {
			if existing.Name == name {
				config.Clusters[i] = newCluster
				exists = true
				break
			}
		}
		if !exists {
			config.Clusters = append(config.Clusters, newCluster)
		}
	}

	// Save updated config as JSON
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Successfully imported %d clusters from kubeconfig\nRemember to `source <(k rc)` for it to take effect.", len(kubeconfig.Clusters))
	return nil
}

func getKubeConfigPath() string {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	return kubeconfigPath
}
