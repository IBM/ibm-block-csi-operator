package utils

import (
	"github.com/IBM/volume-group-operator/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func areLabelsMatchLabelSelector(client client.Client, labelsToCheck map[string]string,
	labelSelector metav1.LabelSelector) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return false, &errors.MatchingLabelsAndLabelSelectorError{ErrorMessage: err.Error()}
	}
	return isSelectorMatchesLabels(selector, labelsToCheck), nil
}

func isSelectorMatchesLabels(selector labels.Selector, labelsToCheck map[string]string) bool {
	set := labels.Set(labelsToCheck)
	return selector.Matches(set)
}
