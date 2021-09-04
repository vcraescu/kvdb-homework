package env

import (
	"fmt"
	"os"
)

func Require(s string) (string, error) {
	v := os.Getenv(s)
	if v == "" {
		return "", fmt.Errorf("%q is missing", s)
	}

	return v, nil
}
