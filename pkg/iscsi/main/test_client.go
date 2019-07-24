package main

import (
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/iscsi/client"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("iscsi agent")

func main() {
	logf.SetLogger(zap.Logger())
	c := client.NewIscsiClient("9.115.241.201:10086", log)
	c.Login([]string{"9.115.241.215", "9.115.241.219"})
	c.Logout([]string{"9.115.241.215", "9.115.241.219"})
}
