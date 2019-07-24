package main

import (
	"flag"
	"os"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/iscsi/server"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("iscsi agent")

func main() {

	address := flag.String("address", "", "Listening Address")
	flag.Parse()

	logf.SetLogger(zap.Logger())

	var addr string = *address
	if addr == "" {
		addr = os.Getenv("ADDRESS")
		if addr == "" {
			log.Error(nil, "--address or env ADDRESS is required!")
			os.Exit(1)
		}
	}

	log.Info("Start server")
	if err := server.Serve(addr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}
