package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	driver "github.com/stakater/k8s-cost-optimizer/pkg/driver"
)

func main() {

	var (
		dryRun         = flag.Bool("dry-run", false, "Only do a dry Run (default: false)")
		tolerance      = flag.Int("tolerance", 0, "Ignore certain weight difference (default: 0)")
		patchResources = flag.Bool("patch", false, "Path resources according to config (default: false)")
		configFilePath = flag.String("config-file-path", "/app/config", "Path to config file where the details of deployments needs to be read")
	)
	flag.Parse()

	err := driver.Drive(*tolerance, *patchResources, *configFilePath, *dryRun)
	if err != nil {
		panic(err)
	}
	logrus.Info("ALL DONE!!!")
}
