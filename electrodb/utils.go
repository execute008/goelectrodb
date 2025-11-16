package electrodb

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// int32Ptr returns a pointer to the given int32
func int32Ptr(i int32) *int32 {
	return &i
}

// boolPtr returns a pointer to the given bool
func boolPtr(b bool) *bool {
	return &b
}
