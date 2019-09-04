// +build ignore

package main

import (
	"log"

	"github.com/IBM/ibm-block-csi-operator/pkg/resources"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(resources.Deploy, vfsgen.Options{
		PackageName:  "resources",
		BuildTags:    "release",
		VariableName: "Deploy",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
