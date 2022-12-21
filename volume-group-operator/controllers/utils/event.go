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
	event := generateVGEvent(vg, reason, message, normalEventType)
	logger.Info(fmt.Sprintf(messages.CreateEventForVolumeGroup, vg.Name, vg.Namespace, message))
	return createEvent(logger, client, event)
}

func createVolumeGroupErrorEvent(logger logr.Logger, client client.Client, vg *volumegroupv1.VolumeGroup,
	errorMessage, reason string) error {
	event := generateVGEvent(vg, reason, errorMessage, warningEventType)
	logger.Info(fmt.Sprintf(messages.CreateEventForVolumeGroup, vg.Name, vg.Namespace, errorMessage))
	return createEvent(logger, client, event)
}

func generateVGEvent(vg *volumegroupv1.VolumeGroup, reason, message, eventType string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: vg.Namespace,
			Name:      fmt.Sprintf("%s.%s", vg.Name, generateString()),
		},
		ReportingController: volumeGroupController,
		InvolvedObject: corev1.ObjectReference{
			Kind:       vg.Kind,
			APIVersion: vg.APIVersion,
			Name:       vg.Name,
			Namespace:  vg.Namespace,
			UID:        vg.UID,
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
