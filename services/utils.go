package services

import (
	"strings"
)

func GetResourcePhysicalIdFromArn(resourceArn string) string {

	resourceArnSlice := strings.Split(resourceArn, "/")

	resourcePhysicalId := resourceArnSlice[len(resourceArnSlice)-1]

	return resourcePhysicalId

}
