package ibmblockcsi

// Validate checks if the spec is validated
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#validation
func (c *IBMBlockCSI) Validate() error {
	return nil
}
