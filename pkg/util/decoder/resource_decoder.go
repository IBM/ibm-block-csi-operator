package decoder

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func FromJsonToUnstructured(json []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	err := obj.UnmarshalJSON(json)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func FromYamlToUnstructured(yaml []byte) (*unstructured.Unstructured, error) {
	json, err := yamlutil.ToJSON(yaml)
	if err != nil {
		return nil, err
	}
	return FromJsonToUnstructured(json)
}

func FromJsonToUnstructuredList(json []byte) (*unstructured.UnstructuredList, error) {
	obj := &unstructured.UnstructuredList{}
	err := obj.UnmarshalJSON(json)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func FromYamlToUnstructuredList(yaml []byte) (*unstructured.UnstructuredList, error) {
	json, err := yamlutil.ToJSON(yaml)
	if err != nil {
		return nil, err
	}
	return FromJsonToUnstructuredList(json)
}
