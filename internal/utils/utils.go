package utils

import (
	"github.com/go-faker/faker/v4"
)

func GenerateFakeName() string {
	name := faker.Name()
	return name
}
