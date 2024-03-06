package utils

import (
	"fmt"
	"math/rand"
)

func GenerateRandCode() string {
	return fmt.Sprintf("%6d", rand.Intn(999999))
}
