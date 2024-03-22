package test

import (
	"fmt"
	"os"
	"testing"
)

func TestEnvVariable(t *testing.T) {

	fmt.Println(os.Getenv("test"))

}
