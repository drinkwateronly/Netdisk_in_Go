package common

import uuid "github.com/satori/go.uuid"

func GenerateUUID() string {
	v4 := uuid.NewV4().String()
	return v4
}
