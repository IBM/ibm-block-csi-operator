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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/node/server"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("node agent")

func main() {

	address := flag.String("address", "", "Listening Address")
	flag.Parse()

	logf.SetLogger(zap.Logger())

	var addr string = *address
	if addr == "" {
		addr = os.Getenv("ADDRESS")
		if addr == "" {
			log.Error(nil, "--address or ENV ADDRESS is required!")
			os.Exit(1)
		}
	}

	nodeName := os.Getenv(config.ENVNodeName)
	if nodeName == "" {
		log.Error(fmt.Errorf("ENV %s is not set", config.ENVNodeName), "")
		os.Exit(1)
	}

	log.Info("Start server", "node", nodeName)
	if err := server.Serve(addr, nodeName); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}
