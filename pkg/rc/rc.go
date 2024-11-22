package rc

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/KevinWang15/k/pkg/model"
	"github.com/KevinWang15/k/pkg/utils"
	"github.com/lithammer/dedent"
	"k8s.io/client-go/tools/clientcmd/api"
)

func Run() {

	utils.EnsureKHomeDir()

	config := utils.GetConfig()

	clusters := config.Clusters

	for _, cluster := range clusters {
		generateKubeConfigForCluster(cluster)
	}

	fmt.Printf(dedent.Dedent(`

function kubectl-k() {
    if [ "$1" = "touch" ]; then
        set -- "annotate" "${@:2}" "touch=$RANDOM$RANDOM" "--overwrite" 
    fi

    if [[ "$@" != *"--all-namespaces"* && "$@" != *"--namespace"* && "$@" != *"-n"* && -n "$K_DEFAULT_NAMESPACE" ]]; then
        set -- -n "$K_DEFAULT_NAMESPACE" "$@"
    fi
	KUBECONFIG=` + path.Join(consts.K_HOME_DIR, "kubeconfigs", "$K_CLUSTER") + ` kubectl $@
}

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

	for _, cluster := range clusters {
		fmt.Printf(dedent.Dedent(`
alias k` + cluster.Name + `='K_CLUSTER=` + cluster.Name + ` kubectl-k'
`))

		for shortcut, expanded := range config.Shortcuts {
			fmt.Printf(dedent.Dedent(`
alias k` + cluster.Name + shortcut + `='K_CLUSTER=` + cluster.Name + ` kubectl-k ` + expanded + `'
`))
		}
	}

}

func generateKubeConfigForCluster(cluster model.Cluster) {
	dir := path.Join(consts.K_HOME_DIR, "kubeconfigs")

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("create dir %q error: %s", dir, err.Error()))
	}

	filePath := path.Join(dir, cluster.Name)

	kubeconfig := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Contexts: map[string]*api.Context{
			"default": {
				Cluster:  "default",
				AuthInfo: "default",
			},
		},
		CurrentContext: "default",
	}

	kubeconfig.Clusters = map[string]*api.Cluster{
		"default": cluster.Cluster.ToAPICluster(),
	}

	kubeconfig.AuthInfos = map[string]*api.AuthInfo{
		"default": cluster.User.ToAPIAuthInfo(),
	}

	bytes, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		panic(fmt.Errorf("write kubeconfig error: %s", err.Error()))
	}

	err = os.WriteFile(filePath, bytes, 0600)

	if err != nil {
		panic(fmt.Errorf("write file %s error: %s", filePath, err.Error()))
	}
}
