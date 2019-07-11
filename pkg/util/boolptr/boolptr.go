package boolptr

// IsTrue returns true if and only if the bool pointer is non-nil and set to true.
func IsTrue(b *bool) bool {
	return b != nil && *b == true
}

// IsFalse returns true if and only if the bool pointer is non-nil and set to false.
func IsFalse(b *bool) bool {
	return b != nil && *b == false
}

// True returns a *bool whose underlying value is true.
func True() *bool {
	t := true
	return &t
}

// False returns a *bool whose underlying value is false.
func False() *bool {
	t := false
	return &t
}

// Equal returns true if and only if both values are set and equal.
func Equal(a, b *bool) bool {
	if a == nil || b == nil {
		return false
	} else {
		return *a == *b
	}
}

// Equal returns true if both values are set and equal or both are nil.
func NilOrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	} else {
		return *a == *b
	}
}
