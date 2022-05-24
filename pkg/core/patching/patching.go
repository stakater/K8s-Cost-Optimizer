package patching

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stakater/k8s-cost-optimizer/pkg/common"
	"github.com/stakater/k8s-cost-optimizer/pkg/types"
	utils "github.com/stakater/k8s-cost-optimizer/pkg/util"
	"gopkg.in/yaml.v2"
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

func PatchResources(client *clientset.Clientset, configFilePath string, dryRun bool) error {
	// Verify config
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		logrus.Errorf("Couldn't get config file at the path: %s", err)
		return err
	}
	buf, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	patchConfig := types.KCOConfig{}
	err = yaml.UnmarshalStrict(buf, &patchConfig)
	if err != nil {
		logrus.Errorf("Couldn't parse config, error: %v", err)
		return err
	}
	patchHash, err := utils.AsSha256(patchConfig)
	if err != nil {
		logrus.Errorf("Couldn't generate hash for config, error: %v", err)
		return err
	}

	// pre prep
	patch := patchConfig.SpecPatch
	var toIgnoreDeployments map[string]bool = make(map[string]bool)
	var toIgnoreStatefulSets map[string]bool = make(map[string]bool)
	var includedNamespaces map[string]bool = make(map[string]bool)
	for _, dep := range patchConfig.ResourcesToIgnore.Deployments {
		toIgnoreDeployments[fmt.Sprintf("%s-%s", dep.Namespace, dep.Name)] = true
	}
	for _, sset := range patchConfig.ResourcesToIgnore.StatefulSets {
		toIgnoreStatefulSets[fmt.Sprintf("%s-%s", sset.Namespace, sset.Name)] = true
	}

	// Remove Patches

	_ = RemovePatch(client, patchConfig, dryRun, toIgnoreStatefulSets, toIgnoreDeployments, includedNamespaces)

	// Check for patching

	for _, namespace := range patchConfig.TargetNamespaces {
		// Deployments
		deps, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, dep := range deps.Items {
			if toIgnoreDeployments[fmt.Sprintf("%s-%s", dep.Namespace, dep.Name)] {
				logrus.Infof("[SKIPPED] deployment %s/%s", dep.Namespace, dep.Name)
				continue
			}
			if dep.ObjectMeta.Annotations == nil {
				dep.ObjectMeta.Annotations = make(map[string]string)
			}
			depPatchHash, ok := dep.ObjectMeta.Annotations[common.KCO_LABLE_KEY_NAME]
			if ok && depPatchHash == patchHash {
				logrus.Infof("[ALREADY PATCHED-IGNORED] deployment %s/%s", dep.Namespace, dep.Name)
				continue
			} else {
				dep.ObjectMeta.Annotations[common.KCO_LABLE_KEY_NAME] = patchHash
				// Add tolerations
				tolerations := []v1.Toleration{}
				for _, tolerance := range patch.Tolerations {
					tolerations = append(tolerations, v1.Toleration{
						Key:      tolerance.Key,
						Operator: v1.TolerationOperator(tolerance.Operator),
						Value:    tolerance.Value,
						Effect:   v1.TaintEffect(tolerance.Effect),
					})
				}
				if dep.Spec.Template.Spec.Tolerations == nil {
					dep.Spec.Template.Spec.Tolerations = []v1.Toleration{}
				}
				dep.Spec.Template.Spec.Tolerations = append(dep.Spec.Template.Spec.Tolerations, tolerations...)
				// Add preferred scheduling values if nil
				if dep.Spec.Template.Spec.Affinity == nil || dep.Spec.Template.Spec.Affinity.NodeAffinity == nil || dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
					dep.Spec.Template.Spec.Affinity = &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{},
						},
					}
				}
				// Add tolerations
				affinities := []v1.PreferredSchedulingTerm{}
				for _, pref := range patch.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					temp := v1.PreferredSchedulingTerm{
						Weight:     pref.Weight,
						Preference: v1.NodeSelectorTerm{},
					}
					for _, pp := range pref.Preference.MatchExpressions {
						temp.Preference.MatchExpressions = append(temp.Preference.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      pp.Key,
							Operator: v1.NodeSelectorOperator(pp.Operator),
							Values:   pp.Values,
						})
					}
					affinities = append(affinities, temp)
				}
				dep.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = affinities
				if !dryRun {
					_, err := client.AppsV1().Deployments(dep.Namespace).Update(context.Background(), &dep, metav1.UpdateOptions{})
					if err != nil {
						logrus.Errorf("[FAILED UPDATE] deployment %s/%s : %s", dep.Namespace, dep.Name, err)
					} else {
						logrus.Infof("[PATCHED] deployment %s/%s", dep.Namespace, dep.Name)
					}
				} else {
					logrus.Infof("[DRYRUN-PATCHED] deployment %s/%s", dep.Namespace, dep.Name)
				}
			}
		}
		// StatefulSets
		ssets, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, sset := range ssets.Items {
			if toIgnoreStatefulSets[fmt.Sprintf("%s-%s", sset.Namespace, sset.Name)] {
				logrus.Infof("[SKIPPED] statefulset %s/%s", sset.Namespace, sset.Name)
				continue
			}

			if sset.ObjectMeta.Annotations == nil {
				sset.ObjectMeta.Annotations = make(map[string]string)
			}
			ssetPatchHash, ok := sset.ObjectMeta.Annotations[common.KCO_LABLE_KEY_NAME]
			if ok && ssetPatchHash == patchHash {
				logrus.Infof("[ALREADY PATCHED-IGNORED] statefulset %s/%s", sset.Namespace, sset.Name)
				continue
			} else {
				sset.ObjectMeta.Annotations[common.KCO_LABLE_KEY_NAME] = patchHash
				// Add tolerations
				tolerations := []v1.Toleration{}
				for _, tolerance := range patch.Tolerations {
					tolerations = append(tolerations, v1.Toleration{
						Key:      tolerance.Key,
						Operator: v1.TolerationOperator(tolerance.Operator),
						Value:    tolerance.Value,
						Effect:   v1.TaintEffect(tolerance.Effect),
					})
				}
				if sset.Spec.Template.Spec.Tolerations == nil {
					sset.Spec.Template.Spec.Tolerations = []v1.Toleration{}
				}
				sset.Spec.Template.Spec.Tolerations = append(sset.Spec.Template.Spec.Tolerations, tolerations...)
				// Add preferred scheduling values if nil
				if sset.Spec.Template.Spec.Affinity == nil || sset.Spec.Template.Spec.Affinity.NodeAffinity == nil || sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
					sset.Spec.Template.Spec.Affinity = &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{},
						},
					}
				}
				// Add tolerations
				affinities := []v1.PreferredSchedulingTerm{}
				for _, pref := range patch.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					temp := v1.PreferredSchedulingTerm{
						Weight:     pref.Weight,
						Preference: v1.NodeSelectorTerm{},
					}
					for _, pp := range pref.Preference.MatchExpressions {
						temp.Preference.MatchExpressions = append(temp.Preference.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      pp.Key,
							Operator: v1.NodeSelectorOperator(pp.Operator),
							Values:   pp.Values,
						})
					}
					affinities = append(affinities, temp)
				}
				sset.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = affinities
				if !dryRun {
					_, err := client.AppsV1().StatefulSets(sset.Namespace).Update(context.Background(), &sset, metav1.UpdateOptions{})
					if err != nil {
						logrus.Errorf("[FAILED UPDATE] statefulset %s/%s : %s", sset.Namespace, sset.Name, err)
					} else {
						logrus.Infof("[PATCHED] statefulset %s/%s", sset.Namespace, sset.Name)
					}
				} else {
					logrus.Infof("[DRYRUN-PATCHED] statefulset %s/%s", sset.Namespace, sset.Name)
				}
			}
		}
	}
	return nil
}
