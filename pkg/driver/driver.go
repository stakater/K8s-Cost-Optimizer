package driver

import (
	"github.com/sirupsen/logrus"
	patch "github.com/stakater/k8s-cost-optimizer/pkg/core/patching"
	rescheduler "github.com/stakater/k8s-cost-optimizer/pkg/core/rescheduler"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Drive(tolerance int, patchResources bool, configFilePath string, dryRun bool) error {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		logrus.Errorf("Couldn't get Kubernetes default config: %s", err)
		return err
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}

	// Check for patching
	if patchResources {
		err = patch.PatchResources(client, configFilePath, dryRun)
		if err != nil {
			return err
		}
		logrus.Info("patching done")
	} else {
		logrus.Info("patching skipped")
	}

	err = rescheduler.CheckReschedulePotentialAndDeleteWorkload(client, tolerance, dryRun)
	if err != nil {
		return err
	}
	return nil
}
