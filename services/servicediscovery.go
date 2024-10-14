package services

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
)

// const DEREGISTER_INSTANCE_MAX_RETRIES = 5

func DeregisterTaskFromCloudMapService(client servicediscovery.Client, taskARN string, discoveryARN string) bool {

	taskARNSlice := strings.Split(taskARN, "/")
	cloudMapServiceIdSlice := strings.Split(discoveryARN, "/")

	taskId := taskARNSlice[len(taskARNSlice)-1]
	cloudMapServiceId := cloudMapServiceIdSlice[len(cloudMapServiceIdSlice)-1]

	var cloudMapServiceInstanceParams servicediscovery.DeregisterInstanceInput
	cloudMapServiceInstanceParams.InstanceId = &taskId
	cloudMapServiceInstanceParams.ServiceId = &cloudMapServiceId

	deregistrationOutput, err := client.DeregisterInstance(context.TODO(), &cloudMapServiceInstanceParams)
	if err != nil {
		log.Fatalf("Failed to deregister %s instance from %s Cloud Map service...", taskId, cloudMapServiceId)
	}

	var cloudMapOperationParams servicediscovery.GetOperationInput
	cloudMapOperationParams.OperationId = deregistrationOutput.OperationId

	operationOutput, err := client.GetOperation(context.TODO(), &cloudMapOperationParams)
	if err != nil {
		log.Fatalf("There was an error getting info for Cloud Map operation ID %s", *deregistrationOutput.OperationId)

	}

	log.Printf("Operation ID %s - Successfully deregistered %s instance from %s Cloud Map service!",
		*operationOutput.Operation.Id,
		taskId,
		cloudMapServiceId,
	)

	return true
	// for i := DEREGISTER_INSTANCE_MAX_RETRIES; i > 0; i-- {

	// 	if i == 0 {
	// 		return false
	// 	}

	// 	operationOutput, err := client.GetOperation(context.TODO(), &cloudMapOperationParams)
	// 	if err != nil {
	// 		log.Fatalf("There was an error getting info for Cloud Map operation ID %s", *deregistrationOutput.OperationId)

	// 	}

	// 	if operationOutput.Operation.Status != "SUCCESS" {
	// 		log.Printf("Checking de-registering instance operation %s - %s of %s",
	// 			*operationOutput.Operation.Id,
	// 			i,
	// 			DEREGISTER_INSTANCE_MAX_RETRIES,
	// 		)
	// 		time.Sleep(time.Duration(time.Second * 2))
	// 		continue
	// 	} else {

	// 		log.Printf("Operation ID %s - Successfully deregistered %s instance from %s Cloud Map service!",
	// 			*operationOutput.Operation.Id,
	// 			taskId,
	// 			cloudMapServiceId,
	// 		)

	// 		return true
	// 	}
	// }
}
