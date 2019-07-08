package kubernetes

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	version "github.com/IBM/ibm-block-csi-driver-operator/version"
)

func ServerVersion(client discovery.DiscoveryInterface) (string, error) {
	versionInfo, err := client.ServerVersion()
	if err != nil {
		return "", errors.Wrap(err, "error getting server version")
	}

	return fmt.Sprintf("%s.%s", versionInfo.Major, versionInfo.Minor), nil
}

// Config returns a *rest.Config, using either the default kubeconfig or an in-cluster configuration.
func KubeConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = ""
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: ""}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	clientConfig.UserAgent = buildUserAgent(
		"ibm-block-csi",
		version.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	return clientConfig, nil
}

// buildUserAgent builds a User-Agent string from given args.
func buildUserAgent(command, version, os, arch string) string {
	return fmt.Sprintf(
		"%s/%s (%s/%s)", command, version, os, arch)
}
