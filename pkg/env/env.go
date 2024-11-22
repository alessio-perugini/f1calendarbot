package env

import (
	"fmt"
	"os"
)

func MustGet(e string) string {
	v := os.Getenv(e)
	if v == "" {
		panic(fmt.Sprintf("env %v not set", e))
	}
	return v
}
