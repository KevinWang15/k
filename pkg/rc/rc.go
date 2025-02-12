package rc

import (
	"fmt"
	"os"
	"path"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/KevinWang15/k/pkg/model"
	"github.com/KevinWang15/k/pkg/utils"
	"github.com/lithammer/dedent"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Run is invoked by `k rc`. It prints shell function definitions and aliases
// that let you use per-cluster shortcuts. We now keep all clusters in a single
// kubeconfig file, with one context per cluster.
func Run() {

	utils.EnsureKHomeDir()
	config := utils.GetConfig()
	clusters := config.Clusters

	// We’ll store the single merged kubeconfig at ~/.k/config
	// and keep all caches under ~/.k/cache
	configPath := path.Join(consts.K_HOME_DIR, "config")
	cacheDir := path.Join(consts.K_HOME_DIR, "cache")

	err := os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("failed to create cache directory %q: %w", cacheDir, err))
	}

	// Build one kubeconfig that includes all clusters
	kubeConfig := generateSingleKubeconfig(clusters)

	// Write that kubeconfig to ~/.k/config
	err = writeKubeconfigToFile(kubeConfig, configPath)
	if err != nil {
		panic(err)
	}

	// Print the shell function: it sets KUBECONFIG to our single config file
	// and always uses --cache-dir=... but defers context selection to the alias.
	fmt.Printf(dedent.Dedent(`

function kubectl-k() {
    if [ "$1" = "touch" ]; then
        set -- "annotate" "${@:2}" "touch=$RANDOM$RANDOM" "--overwrite"
    fi

    # Default namespace injection if none specified
    if [[ "$@" != *"--all-namespaces"* && "$@" != *"--namespace"* && "$@" != *"-n"* && -n "$K_DEFAULT_NAMESPACE" ]]; then
        set -- -n "$K_DEFAULT_NAMESPACE" "$@"
    fi

      KUBECONFIG=`+configPath+` kubectl --cache-dir='`+cacheDir) + `' $@
}

`)

	// Helper for changing the default namespace
	fmt.Printf(dedent.Dedent(`

function kns() {
    if [ -n "$1" ]; then
        export K_DEFAULT_NAMESPACE="$1"
    else
        echo "Warn: No namespace provided, namespace override disabled"
        export K_DEFAULT_NAMESPACE=""
    fi
}

function watch-changes() {
    cmdToRun="$(alias $1 | awk -F\' '{print $2}')"
    shift
    cmdToRun="$cmdToRun $@"
    cmdToRun="$cmdToRun -ojson --output-watch-events --watch"
    cmdToRun="while true; do $cmdToRun || break; done | k watch-changes"

    eval "$cmdToRun"
}

`))

	// For each cluster, create an alias that sets --context=<clusterName>.
	// Also create aliases for shortcuts.
	for _, cluster := range clusters {
		fmt.Printf("alias k%s='kubectl-k --context %s'\n", cluster.Name, cluster.Name)

		for shortcut, expanded := range config.Shortcuts {
			fmt.Printf("alias k%v%v='kubectl-k --context %v %v'\n", cluster.Name, shortcut, cluster.Name, expanded)
		}
	}
}

// generateSingleKubeconfig constructs a single api.Config with multiple contexts, clusters, and users.
func generateSingleKubeconfig(clusters []model.Cluster) *api.Config {
	kcfg := &api.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       make(map[string]*api.Cluster),
		AuthInfos:      make(map[string]*api.AuthInfo),
		Contexts:       make(map[string]*api.Context),
		CurrentContext: "",
	}

	// If we have at least one cluster, set the first as CurrentContext
	if len(clusters) > 0 {
		kcfg.CurrentContext = clusters[0].Name
	}

	for _, c := range clusters {
		clusterAPI := c.Cluster.ToAPICluster()
		userAPI := c.User.ToAPIAuthInfo()

		// Use the cluster's name as the key in each map
		kcfg.Clusters[c.Name] = clusterAPI
		kcfg.AuthInfos[c.Name] = userAPI
		kcfg.Contexts[c.Name] = &api.Context{
			Cluster:  c.Name,
			AuthInfo: c.Name,
		}
	}
	return kcfg
}

// writeKubeconfigToFile serializes the api.Config to YAML and writes it to path.
func writeKubeconfigToFile(kcfg *api.Config, filePath string) error {
	bytes, err := clientcmd.Write(*kcfg)
	if err != nil {
		return fmt.Errorf("failed to marshal merged kubeconfig: %w", err)
	}

	err = os.WriteFile(filePath, bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write merged kubeconfig to %s: %w", filePath, err)
	}
	return nil
}
