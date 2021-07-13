/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	version "github.com/IBM/ibm-block-csi-operator/version"
)

var log = logf.Log.WithName("kube_util")
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
