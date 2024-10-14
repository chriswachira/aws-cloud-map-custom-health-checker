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
	Cluster       string
	TaskARN       string
	Family        string
	Revision      string
	DesiredStatus string
	KnownStatus   string
}

func GetECSServiceForTask(ecsTask types.Task) string {
	ecsServiceName := strings.TrimPrefix(*ecsTask.Group, "service:")

	log.Printf("Task %s belongs in the %s service", *ecsTask.TaskArn, ecsServiceName)

	return ecsServiceName
}

func GetServiceConnectResources(client *ecs.Client, cluster, ecsService string) (types.ServiceConnectServiceResource, bool) {

	var serviceParams ecs.DescribeServicesInput
	serviceParams.Cluster = &cluster
	serviceParams.Services = []string{ecsService}

	output, err := client.DescribeServices(context.TODO(), &serviceParams)

	if err != nil {
		log.Fatal("There was an error when describing the services: ", err)
	}

	log.Println("Successfully fetched information for service -", ecsService)
	latestDeployment := output.Services[0].Deployments[0]

	if latestDeployment.ServiceConnectConfiguration == nil {
		log.Printf("Service Connect is not enabled for the %s service! Exiting...", ecsService)

		var dummyServiceConnectResources types.ServiceConnectServiceResource
		return dummyServiceConnectResources, false
	}

	log.Printf("Service Connect is enabled for the %s service! Proceeding...", ecsService)
	return latestDeployment.ServiceConnectResources[0], true

}

func GetTaskV4Metadata(taskMetadataEndpoint string) FargateTaskMetadataV4Response {
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

	return string(ecsTask.HealthStatus)

}

func DescribeTask(client ecs.Client, taskMetadataResp FargateTaskMetadataV4Response) types.Task {

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
