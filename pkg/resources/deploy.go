// +build !release
//go:generate go run generate.go

package resources

import (
	"net/http"
)

// Deploy contains yaml files for deployment.
var Deploy http.FileSystem = http.Dir("../../deploy")
