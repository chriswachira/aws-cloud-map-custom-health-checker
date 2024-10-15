package services

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
)

// const DEREGISTER_INSTANCE_MAX_RETRIES = 5

// func UpdateInstanceCustomHealthStatus(client servicediscovery.Client, taskArn string, discoveryArn string, status types.CustomHealthStatus) {
// 	taskARNSlice := strings.Split(taskArn, "/")
// 	cloudMapServiceIdSlice := strings.Split(discoveryArn, "/")

// 	taskId := taskARNSlice[len(taskARNSlice)-1]
// 	cloudMapServiceId := cloudMapServiceIdSlice[len(cloudMapServiceIdSlice)-1]

// 	log.Println("TaskID - ", taskId)
// 	log.Println("Service ID - ", cloudMapServiceId)

// 	var instanceCustomHealthStatusParams servicediscovery.UpdateInstanceCustomHealthStatusInput
// 	instanceCustomHealthStatusParams.InstanceId = &taskId
// 	instanceCustomHealthStatusParams.ServiceId = &cloudMapServiceId
// 	instanceCustomHealthStatusParams.Status = status

// 	_, err := client.UpdateInstanceCustomHealthStatus(context.TODO(), &instanceCustomHealthStatusParams)
// 	if err != nil {
// 		log.Fatal("There was an error updating the instance's health to Cloud Map service: ", err)
// 	}

// 	log.Printf("Successfully updated health status of instance %s to %s on %s!", taskId, string(status), cloudMapServiceId)

// }

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
