package services

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
)

func DeregisterTaskFromCloudMapService(client servicediscovery.Client, taskId string, cloudMapServiceId string) bool {

	var cloudMapServiceInstanceParams servicediscovery.DeregisterInstanceInput
	cloudMapServiceInstanceParams.InstanceId = &taskId
	cloudMapServiceInstanceParams.ServiceId = &cloudMapServiceId

	deregistrationOutput, err := client.DeregisterInstance(context.TODO(), &cloudMapServiceInstanceParams)
	if err != nil {
		log.Fatalf("Failed to deregister %s instance from %s Cloud Map service: %s ", taskId, cloudMapServiceId, err)
	}

	var cloudMapOperationParams servicediscovery.GetOperationInput
	cloudMapOperationParams.OperationId = deregistrationOutput.OperationId

	operationOutput, err := client.GetOperation(context.TODO(), &cloudMapOperationParams)
	if err != nil {
		log.Fatalf("There was an error getting info for Cloud Map operation ID %s: %s", *deregistrationOutput.OperationId, err)

	}

	log.Printf("Operation ID %s - Successfully deregistered %s instance from %s Cloud Map service!",
		*operationOutput.Operation.Id,
		taskId,
		cloudMapServiceId,
	)

	return true

}
