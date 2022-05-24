package patching

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stakater/k8s-cost-optimizer/pkg/common"
	"github.com/stakater/k8s-cost-optimizer/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func RemovePatch(client *clientset.Clientset, patchConfig types.KCOConfig, dryRun bool, toIgnoreStatefulSets map[string]bool, toIgnoreDeployments map[string]bool, includedNamespaces map[string]bool) error {
	for _, namespace := range patchConfig.TargetNamespaces { // this will help in less iterations in next loop
		includedNamespaces[namespace] = true
	}
	patch := patchConfig.SpecPatch
	deps, err := client.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	removePatchDeploymentFunction := func(dep appsv1.Deployment) {
		delete(dep.ObjectMeta.Annotations, common.KCO_LABLE_KEY_NAME)
		// only 1 item as per bounds
		toMatch := patch.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
		toRemoveIndex := -1
		for i, v := range dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			if v.Weight == toMatch.Weight &&
				v.Preference.MatchExpressions[0].Key == toMatch.Preference.MatchExpressions[0].Key &&
				v.Preference.MatchExpressions[0].Operator == v1.NodeSelectorOperator(toMatch.Preference.MatchExpressions[0].Operator) &&
				v.Preference.MatchExpressions[0].Values[0] == toMatch.Preference.MatchExpressions[0].Values[0] {
				toRemoveIndex = i
			}
		}
		dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[:toRemoveIndex], dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[toRemoveIndex+1:]...)

		toMatchTol := patch.Tolerations[0]
		for i, v := range dep.Spec.Template.Spec.Tolerations {
			if toMatchTol.Key == v.Key &&
				toMatchTol.Operator == string(v.Operator) &&
				toMatchTol.Value == v.Value &&
				toMatchTol.Effect == string(v.Effect) {
				toRemoveIndex = i
			}
		}
		dep.Spec.Template.Spec.Tolerations = append(dep.Spec.Template.Spec.Tolerations[:toRemoveIndex], dep.Spec.Template.Spec.Tolerations[toRemoveIndex+1:]...)

		if !dryRun {
			_, err := client.AppsV1().Deployments(dep.Namespace).Update(context.Background(), &dep, metav1.UpdateOptions{})
			if err != nil {
				logrus.Errorf("[FAILED PATCH-REMOVED] deployment %s/%s : %s", dep.Namespace, dep.Name, err)
			} else {
				logrus.Infof("[PATCH-REMOVED] deployment %s/%s", dep.Namespace, dep.Name)
			}
		} else {
			logrus.Infof("[DRYRUN-PATCH-REMOVED] deployment %s/%s", dep.Namespace, dep.Name)
		}
	}

	removePatchStatefulFunction := func(sset appsv1.StatefulSet) {
		delete(sset.ObjectMeta.Annotations, common.KCO_LABLE_KEY_NAME)
		// only 1 item as per bounds
		toMatch := patch.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
		toRemoveIndex := -1
		for i, v := range sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			if v.Weight == toMatch.Weight &&
				v.Preference.MatchExpressions[0].Key == toMatch.Preference.MatchExpressions[0].Key &&
				v.Preference.MatchExpressions[0].Operator == v1.NodeSelectorOperator(toMatch.Preference.MatchExpressions[0].Operator) &&
				v.Preference.MatchExpressions[0].Values[0] == toMatch.Preference.MatchExpressions[0].Values[0] {
				toRemoveIndex = i
			}
		}
		sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[:toRemoveIndex], sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[toRemoveIndex+1:]...)

		toMatchTol := patch.Tolerations[0]
		for i, v := range sset.Spec.Template.Spec.Tolerations {
			if toMatchTol.Key == v.Key &&
				toMatchTol.Operator == string(v.Operator) &&
				toMatchTol.Value == v.Value &&
				toMatchTol.Effect == string(v.Effect) {
				toRemoveIndex = i
			}
		}
		sset.Spec.Template.Spec.Tolerations = append(sset.Spec.Template.Spec.Tolerations[:toRemoveIndex], sset.Spec.Template.Spec.Tolerations[toRemoveIndex+1:]...)

		if !dryRun {
			_, err := client.AppsV1().StatefulSets(sset.Namespace).Update(context.Background(), &sset, metav1.UpdateOptions{})
			if err != nil {
				logrus.Errorf("[FAILED PATCH-REMOVED] statefulset %s/%s : %s", sset.Namespace, sset.Name, err)
			} else {
				logrus.Infof("[PATCH-REMOVED] statefulset %s/%s", sset.Namespace, sset.Name)
			}
		} else {
			logrus.Infof("[DRYRUN-PATCH-REMOVED] statefulset %s/%s", sset.Namespace, sset.Name)
		}
	}

	for _, dep := range deps.Items {
		_, okNS := includedNamespaces[dep.Namespace] // if NS exists
		_, okIG := toIgnoreDeployments[fmt.Sprintf("%s-%s", dep.Namespace, dep.Name)]
		_, okAN := dep.Annotations[common.KCO_LABLE_KEY_NAME]
		if !okNS && okAN { // If Namespace doesn't exist and annotation exists - remove patch
			removePatchDeploymentFunction(dep)
		}
		if okNS && okIG && okAN { // If Namespace exists and Dep is in ignore and annotation exists - remove patch
			removePatchDeploymentFunction(dep)
		}
	}
	ssets, err := client.AppsV1().StatefulSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sset := range ssets.Items {
		_, okNS := includedNamespaces[sset.Namespace] // if NS exists
		_, okIG := toIgnoreStatefulSets[fmt.Sprintf("%s-%s", sset.Namespace, sset.Name)]
		_, okAN := sset.Annotations[common.KCO_LABLE_KEY_NAME]
		if !okNS && okAN { // If Namespace doesn't exist and annotation exists - remove patch
			// remove patch
			removePatchStatefulFunction(sset)
		}
		if okNS && okIG && okAN { // If Namespace exists and Dep is in ignore and annotation exists - remove patch
			// remove patch
			removePatchStatefulFunction(sset)
		}
	}
	return nil
}
