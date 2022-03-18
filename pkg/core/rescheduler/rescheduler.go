package rescheduler

import (
	"context"

	"github.com/sirupsen/logrus"
	utils "github.com/stakater/k8s-cost-optimizer/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func CheckReschedulePotentialAndDeleteWorkload(client *clientset.Clientset, tolerance int, dryRun bool) error {
	// Check for rescheduling
	hasPotential := false

	nodes, err := utils.ListScheduleableNodes(client)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		pods, err := utils.ListPodsOnNode(client, node)
		if err != nil {
			return err
		}

		for _, pod := range pods {
			curScore, err := utils.CalcPodPriorityScore(pod, node)
			if err != nil {
				return err
			}

			foundBetter, _, nodeNameBetter := utils.FindBetterPreferredNode(pod, curScore, tolerance, nodes)
			if foundBetter {
				hasPotential = true
				logrus.Infof("Pod %v/%v can possibly be scheduled on %v", pod.Namespace, pod.Name, nodeNameBetter)
				if !dryRun {
					err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
					if err != nil {
						return err
					}
					logrus.Infof("Pod %v/%v has been evicted!", pod.Namespace, pod.Name)
				} else {
					logrus.Infof("[DRYRUN] Pod %v/%v has been evicted!", pod.Namespace, pod.Name)
				}
			}
		}
	}

	if !hasPotential {
		logrus.Info("No Pods to evict")
	}
	return nil
}
