package clone

func Map(original map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{}, len(original))
	for k, v := range original {
		switch val := v.(type) {
		case map[string]interface{}:
			clone[k] = Map(val)
		default:
			// For other types, assign directly (if you have slices or other references, add more handling)
			clone[k] = val
		}
	}
	return clone
}
