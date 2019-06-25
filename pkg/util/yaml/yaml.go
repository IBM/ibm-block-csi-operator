package yaml

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// Split splits YAML streams from file into individual documents.
func SplitFromFile(filename string) ([][]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return split(f)
}

// Split splits YAML streams into individual documents.
func Split(data []byte) ([][]byte, error) {
	rc := ioutil.NopCloser(bytes.NewReader(data))
	return split(rc)
}

// If an io.ErrShortBuffer error is returned, means that there is a very
// large yaml resource, you should increase the size of the resource below.
func split(rc io.ReadCloser) ([][]byte, error) {
	decoder := utilyaml.NewDocumentDecoder(rc)
	out := [][]byte{}
	for {
		resource := make([]byte, 6000)
		if _, err := decoder.Read(resource); err != nil {
			decoder.Close()
			if err == io.EOF {
				return out, nil
			} else {
				return nil, err
			}
		}
		out = append(out, resource)
	}
	return out, nil
}
