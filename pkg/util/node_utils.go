package utils

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	clientset "k8s.io/client-go/kubernetes"
)

func IsScheduleable(node *v1.Node) bool {

	for _, taint := range node.Spec.Taints {
		if taint.Key == "node.kubernetes.io/unreachable" {
			return false
		}
	}

	return true
}

func ListScheduleableNodes(client clientset.Interface) ([]*v1.Node, error) {
	nodeList, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []*v1.Node{}, err
	}

	nodes := make([]*v1.Node, 0)
	for i := range nodeList.Items {
		if IsScheduleable(&nodeList.Items[i]) {
			nodes = append(nodes, &nodeList.Items[i])
		}
	}
	return nodes, nil
}

func ListPodsOnNode(client clientset.Interface, node *v1.Node) ([]*v1.Pod, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name + ",status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return []*v1.Pod{}, err
	}

	podList, err := client.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return []*v1.Pod{}, err
	}

	pods := make([]*v1.Pod, 0)
	for i := range podList.Items {
		pods = append(pods, &podList.Items[i])
	}
	return pods, nil
}

func CalcPodPriorityScore(pod *v1.Pod, node *v1.Node) (int, error) {
	var count int32
	affinity := pod.Spec.Affinity
	if affinity != nil && affinity.NodeAffinity != nil && affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		for _, preferredSchedulingTerm := range affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			if preferredSchedulingTerm.Weight == 0 {
				continue
			}

			nodeSelector, err := NodeSelectorRequirementsAsSelector(preferredSchedulingTerm.Preference.MatchExpressions)
			if err != nil {
				return 0, err
			}

			if nodeSelector.Matches(labels.Set(node.Labels)) {
				count += preferredSchedulingTerm.Weight
			}
		}
	}

	return int(count), nil
}

func FindBetterPreferredNode(pod *v1.Pod, curScore int, tolerance int, nodes []*v1.Node) (bool, int, string) {
	for _, node := range nodes {
		score, err := CalcPodPriorityScore(pod, node)
		if err != nil {
			continue
		}

		if (score - tolerance) > curScore {
			return true, score, node.Name
		}
	}

	return false, 0, ""
}

// NodeSelectorRequirementsAsSelector converts the []NodeSelectorRequirement core type into a struct that implements
// labels.Selector.
func NodeSelectorRequirementsAsSelector(nsm []v1.NodeSelectorRequirement) (labels.Selector, error) {
	if len(nsm) == 0 {
		return labels.Nothing(), nil
	}
	selector := labels.NewSelector()
	for _, expr := range nsm {
		var op selection.Operator
		switch expr.Operator {
		case v1.NodeSelectorOpIn:
			op = selection.In
		case v1.NodeSelectorOpNotIn:
			op = selection.NotIn
		case v1.NodeSelectorOpExists:
			op = selection.Exists
		case v1.NodeSelectorOpDoesNotExist:
			op = selection.DoesNotExist
		case v1.NodeSelectorOpGt:
			op = selection.GreaterThan
		case v1.NodeSelectorOpLt:
			op = selection.LessThan
		default:
			return nil, fmt.Errorf("%q is not a valid node selector operator", expr.Operator)
		}
		r, err := labels.NewRequirement(expr.Key, op, expr.Values)
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}
