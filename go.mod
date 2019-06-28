module github.com/IBM/ibm-block-csi-driver-operator

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/go-autorest v11.5.2+incompatible // indirect
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/coreos/prometheus-operator v0.29.0 // indirect
	github.com/dgrijalva/jwt-go v0.0.0-20161101193935-9ed569b5d1ac // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.2
	github.com/go-test/deep v1.0.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20180924190550-6f2cf27854a4 // indirect
	github.com/google/btree v0.0.0-20160524151835-7d79101e329e // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20170728041850-787624de3eb7 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/operator-framework/operator-sdk v0.8.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/presslabs/controller-util v0.1.13
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/api v0.3.2 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190627205229-acea843d18eb
	k8s.io/apiextensions-apiserver v0.0.0-20190627210706-e1f2db1c2108 // indirect
	k8s.io/apimachinery v0.0.0-20190627205106-bc5732d141a8
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190627204931-86aa6a6a5cf3
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.11
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

// Pinned to kubernetes-1.13.1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
)
