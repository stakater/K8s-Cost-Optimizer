# K8s-cost-optimizer

This tool focuses on optimizing cost of kubernetes by taking different use cases into account and making appropriate changes to the cluster.

## Use cases

---

Problem to solve here is to save cost on kubernetes as Kubernetes is usually a multi-node environment so by catering following different scenarios, we can save cost by introducing low cost nodes in our cluster coupled with this tool.

### **1. Workload Balancing**

There are special type of nodes on some cloud providers called spot instances, spot instances are cheap but can be taken away anytime. When this instance is taken back from you by Cloud Provider, but your project can afford some performance degradation for a short period of time (untill there's a new node given), `kube-scheduler` will place it on some other node available at the moment or on a different auto-scalable node set.

Issue here is, when finally original node will join the cluster back, there's nothing to schedule your workload Pods back to it and they still be running on high costed nodes.

Here again **k8s-cost-optimizer** can be useful.

### **2. Workload Patching**

When you need to patch all workloads to be scheduled on a low cost machine. This coupled with **1st** scenario can schedule all the workload on low cost node by patching them first (if not patched) to prefer to schedule on low cost node (specified via config) and then re-deploying them onto other nodes.

## Deploying to Kubernetes

---

You can deploy K8s Cost Optimizer by following methods:

### Vanilla Manifests

You can apply vanilla manifests by changing RELEASE-NAME placeholder provided in manifest with a proper value and apply it by running the command given below:

```bash
kubectl apply -f https://raw.githubusercontent.com/stakater/k8s-cost-optimizer/master/deployments/kubernetes/k8s-cost-optimizer.yaml
```

By default, K8s Cost Optimizer gets deployed in default namespace and watches changes secrets and configmaps in all namespaces.

### Helm Charts

Alternatively if you have configured helm on your cluster, you can add K8s Cost Optimizer to helm from our public chart repository and deploy it via helm using below mentioned commands.

```bash
helm repo add stakater https://stakater.github.io/stakater-charts

helm repo update

helm install stakater/k8scostoptimizer 
```
## How it works

K8s cost optimizer deploys as `CronJob` with run at every 2nd minute by default and it uses a `configMap` and mounts it as it's config.
## Commandline arguments

---

```bash
Usage of k8s-cost-optimizer:
  -config-file-path string
        Path to config file (default "/app/config")
  -dry-run
        Only do a dry Run (default: false)
  -patch
        Path resources according to config (default: false)
  -tolerance int
        Ignore certain weight difference (default: 0)
```

## Config structure

---

Config structure for the moment is a YAML. Which should look like this.
At the moment only `Deployment` & `StatefulSet` type is supported.

Idea behind the config is to patch all the workloads in `targetNamespaces` and ignore the one's that are provided to be ignored in the config YAML.

Mandatory structure for config is:

```YAML
targetNamespaces:
  []
resourcesToIgnore:
  deployments:
    []
  statefuleSets:
    []
  specPatch:
    tolerations:
      []
    affinity:
      nodeAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
          []
```

SAMPLE:

```YAML
targetNamespaces:
- test-ns
resourcesToIgnore:
  deployments:
  - namespace: test
    name: app
  statefuleSets:
  - namespace: test2
    name: app2
specPatch:
  tolerations:
  - effect: NoSchedule
    key: kubernetes.azure.com/scalesetpriority
    operator: Equal
    value: spot
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        preference:
          matchExpressions:
          - key: agentpool
            operator: In
            values:
            - spot
```

## Note

For now, we need to keep the config spec patch static
```
...
specPatch:
  tolerations:
  - effect: NoSchedule
    key: kubernetes.azure.com/scalesetpriority
    operator: Equal
    value: spot
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        preference:
          matchExpressions:
          - key: agentpool
            operator: In
            values:
            - spot
...
```
Reason for this is to remove the patch when and where needed.
This is for the special use-case for our internal operations.