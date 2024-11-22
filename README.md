# k

`k` is an open-source enhancement of `kubectl`, designed to increase your productivity. It offers a variety of features that make Kubernetes cluster management more efficient and intuitive.

> **Note:** This tool is still under active development. Although it has shown significant potential and is being used extensively within the creator's company, further validation and code improvement are anticipated. Contributions are always welcome!

## Installation

To install `k`, use the following command:

```bash
go install .
```

After installation, you need to add the following lines to your `.bashrc` or `.zshrc` file:

```bash
source <(k rc)
```

## Configuration

### Importing configuration from existing KUBECONFIG

You can import your existing Kubernetes configurations from your KUBECONFIG file into the k tool. This will automatically convert all your clusters, including their authentication details, into the k configuration format.
To import your existing configurations:

```bash
k import
```

This will import all your clusters from your KUBECONFIG (default: `~/.kube/config`), including their authentication settings. Existing clusters in your `k` configuration will be updated if they share the same name.

### Configuration

Check out the configuration located at `~/.k/config.json`. 

## Features

### Generating Multiple Kubeconfigs

`k` generates multiple kubeconfigs based on your configuration file. For instance:

* `~/.k/kubeconfigs/l`
* `~/.k/kubeconfigs/l2`

### Aliases and Shortcuts

`k` provides aliases and shortcuts for commands, making them easier and faster to use. For example, if `l` is a cluster, you can use:

```bash
kl get pods
```

Instead of:

```bash
kubectl --kubeconfig ~/.k/kubeconfigs/l get pods
```

You can also define custom shortcuts in the `shortcuts` section of your `~/.k/config.json` file.

For example, if you have `"gp": "get pod"`, then

```
klgp=kubectl --kubeconfig ~/.k/kubeconfigs/l get pods
```

You can define shortcuts in your `~/.k/config.json` file as follows:

```json
{
  "shortcuts": {
    "gp": "get pod",
    "gd": "get deploy",
    "gsvc": "get svc",
    "ci": "cluster-info"
  }
}
```

### Quick Namespace Switching

You can switch namespaces quickly with the following command:

```bash
kns kube-system
```

After switching, 

```
klgp=kubectl --kubeconfig ~/.k/kubeconfigs/l -n kube-system get pods
```

### Watch Kubernetes Resources and Show Diff

Monitor changes in Kubernetes resources and show differences using commands like:

```bash
watch-changes kl get configmap
watch-changes kl get configmap --all-namespaces
watch-changes kl get configmap aaa
```

### Touch

You can trigger a change in a resource with:

```bash
kl touch configmap aaa
```

This is useful when you want to initiate a controller resync.

This command is equivalent to `kl annotate ... "touch=$(date)" --overwrite`

### Scripting Capabilities

You can also use `k` in scripts to perform actions across multiple clusters:

```bash
for cluster in $(k get-all-clusters); do 
  echo "Visiting cluster $cluster"
  K_CLUSTER=$cluster kubectl-k cluster-info; 
done
```

## Future Development

The following features and improvements are planned:

- [ ] Refactor existing code for efficiency and maintainability
- [ ] Develop a comprehensive installation guide
- [ ] Implement functionality to detect mis-installation issues
