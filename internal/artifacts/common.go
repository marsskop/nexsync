package artifacts

import (
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
)

// Helper wrappers for gojsonq
func findString(where *gojsonq.JSONQ, query string) (string, error) {
	str, err := where.FindR(query)
	if err != nil {
		return "", err
	}
	result, _ := str.String()
	return result, nil
}

func pluckStringSlice(where *gojsonq.JSONQ, query string) ([]string, error) {
	str, err := where.PluckR(query)
	if err != nil {
		return []string{}, err
	}
	result, _ := str.StringSlice()
	return result, nil
}

// A componentDict is a useful struct for storing component data
type componentDict struct {
	name, Version, group, path string
	assets                     *gojsonq.JSONQ
}