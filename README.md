﻿# K8s-cost-optimizer

This tool focuses on optimizing cost of kubernetes by taking different use cases into account and making appropriate changes to the cluster.

## Use cases

---

### **1. Workload Balancing**

When a project is small and there's a need to use some special Spot instance node type.
When this instance is taken back from you by AWS, but your project can afford some performance degradation for a short period of time (untill there's a new node given), `kube-scheduler` will place it on some other node available at the moment.
But when finally new Tesla node will join the cluster, there's nothing to schedule your project's Pod back to it.

Here **k8s-cost-optimizer** can be usefull.

### **2. Workload Patching**

When you need to patch all workloads to be scheduled on a low cost machine. This coupled with **1st** use case can schedule all the workload on low cost node.

## Installation

---

```bash
```

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
