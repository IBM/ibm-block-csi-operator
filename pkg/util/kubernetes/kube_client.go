package kubernetes

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("kube_client")
var KubeClient = initKubeClient()

func initKubeClient() *kubernetes.Clientset {
	clientConfig, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	return client
}
