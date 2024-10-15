package services

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type FargateTaskMetadataV4Response struct {

	// Simple struct that will be used in JSON unmarshalling for data from
	// the Task Metadata V4 endpoint.

	Cluster       string
	TaskARN       string
	Family        string
	Revision      string
	DesiredStatus string
	KnownStatus   string
}

func GetECSServiceForTask(ecsTask types.Task) string {

	// This function gets a task struct and returns the service name where
	// the task is a part of.

	ecsServiceName := strings.TrimPrefix(*ecsTask.Group, "service:")

	log.Printf("Task %s belongs in the %s service", *ecsTask.TaskArn, ecsServiceName)

	return ecsServiceName
}

func GetServiceConnectResources(client *ecs.Client, cluster, ecsService string) (types.ServiceConnectServiceResource, bool) {

	// This function returns Service Connect Resources for a given ECS service. The
	// returned information is a ServiceConnectServiceResource type and a boolean indicating
	// where the Service has been configured with Service Connect or not.

	var serviceParams ecs.DescribeServicesInput
	serviceParams.Cluster = &cluster
	serviceParams.Services = []string{ecsService}

	output, err := client.DescribeServices(context.TODO(), &serviceParams)
	if err != nil {
		log.Fatal("There was an error when describing the services: ", err)
	}

	log.Println("Successfully fetched information for service -", ecsService)

	// Get latest deployment for the service.
	latestDeployment := output.Services[0].Deployments[0]

	// Check is the is a ServiceConnectConfiguration, if not present, return a nil type and false
	if latestDeployment.ServiceConnectConfiguration == nil {
		log.Printf("Service Connect is not enabled for the %s service! Exiting...", ecsService)

		var dummyServiceConnectResources types.ServiceConnectServiceResource
		return dummyServiceConnectResources, false
	}

	log.Printf("Service Connect is enabled for the %s service! Proceeding...", ecsService)
	return latestDeployment.ServiceConnectResources[0], true

}

func GetTaskV4Metadata(taskMetadataEndpoint string) FargateTaskMetadataV4Response {

	// This function makes a HTTP request to the ECS Task Metadata V4 endpoint and
	// unmarshalls (decodes) the JSON data into our FargateTaskMetadataV4Response struct.

	taskMetadataResp, err := http.Get(taskMetadataEndpoint + "/task")
	if err != nil {
		log.Fatal("There was an error fetching tasks metadata: ", err)
	}
	defer taskMetadataResp.Body.Close()

	body, err := io.ReadAll(taskMetadataResp.Body)
	if err != nil {
		log.Fatal("There was an error reading HTTP body: ", err)
	}

	var taskMetadataResponse FargateTaskMetadataV4Response
	err = json.Unmarshal(body, &taskMetadataResponse)
	if err != nil {
		log.Fatal("There was an error decoding the JSON data: ", err)
	}

	return taskMetadataResponse
}

func GetTaskHealthStatus(ecsTask types.Task) string {

	// This function gets a given task's health status from a task type.

	return string(ecsTask.HealthStatus)

}

func DescribeTask(client ecs.Client, taskMetadataResp FargateTaskMetadataV4Response) types.Task {

	// This function makes a call to the ECS API for fetching information about a task.
	// Information about a task's service is present in the API's response.

	var tasks = []string{taskMetadataResp.TaskARN}
	var include []types.TaskField
	var taskParams ecs.DescribeTasksInput

	taskParams.Cluster = &taskMetadataResp.Cluster
	taskParams.Tasks = tasks
	taskParams.Include = include

	ecsTaskDetails, err := client.DescribeTasks(context.TODO(), &taskParams)
	if err != nil {
		log.Fatal("There was an error when describing the tasks: ", err)
	}

	return ecsTaskDetails.Tasks[0]

}
