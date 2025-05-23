package containers

import (
	"fmt"
	"strings"
)

func validateContainerName(name string) (string, string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("container name must be specified in the form 'group/container'")
	}
	return parts[0], parts[1], nil
}

func mustContainerName(name string) (string, string) {
	g, c, err := validateContainerName(name)
	if err != nil {
		panic(err.Error())
	}
	return g, c
}
