package utils

import (
	"context"
	"fmt"
	"time"

	volumegroupv1 "github.com/IBM/volume-group-operator/api/v1"
	"github.com/IBM/volume-group-operator/pkg/messages"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createSuccessVolumeGroupEvent(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup,
	message, reason string) error {
	event := generateEvent(vg, reason, message, normalEventType)
	logger.Info(fmt.Sprintf(messages.CreateEventForNamespacedObject, vg.Name, vg.Namespace, vg.Kind, message))
	return createEvent(logger, client, event)
}

func createNamespacedObjectErrorEvent(logger logr.Logger, client client.Client, object client.Object,
	errorMessage, reason string) error {
	event := generateEvent(object, reason, errorMessage, warningEventType)
	logger.Info(fmt.Sprintf(messages.CreateEventForNamespacedObject, object.GetNamespace(), object.GetName(),
		object.GetObjectKind().GroupVersionKind().Kind, errorMessage))
	return createEvent(logger, client, event)
}

func generateEvent(object client.Object, reason, message, eventType string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: object.GetNamespace(),
			Name:      fmt.Sprintf("%s.%s", object.GetName(), generateString()),
		},
		ReportingController: volumeGroupController,
		InvolvedObject: corev1.ObjectReference{
			Kind:       object.GetObjectKind().GroupVersionKind().Kind,
			APIVersion: object.GetObjectKind().GroupVersionKind().Version,
			Name:       object.GetName(),
			Namespace:  object.GetNamespace(),
			UID:        object.GetUID(),
		},
		Reason:  reason,
		Message: message,
		Type:    eventType,
		FirstTimestamp: metav1.Time{
			Time: time.Now(),
		},
	}
}

func createEvent(logger logr.Logger, client client.Client, event *corev1.Event) error {
	err := client.Create(context.TODO(), event)
	if err != nil {
		logger.Error(err, fmt.Sprintf(messages.FailedToCreateEvent, event.Namespace, event.Name))
		return err
	}
	logger.Info(fmt.Sprintf(messages.EventCreated, event.Namespace, event.Name))
	return nil
}
