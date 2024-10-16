package main

import (
	"context"
	"log"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/chriswachira/aws-cloud-map-custom-health-checker/services"
)

// List of task lifecycle states that a task transitions to from RUNNING.
var ecsTaskShutdownStates = []string{
	"DEACTIVATING",
	"DEPROVISIONING",
	"STOPPING",
}

func main() {

	log.Println("Waiting for task to fully initialize; sleeping for 60 seconds...")
	time.Sleep(time.Duration(time.Second * 60))

	log.Println("Initializing AWS Cloud Map Custom Health Checker for Amazon ECS...")

	// Fetch the V4 Metadata URI from the injected environment variable.
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4-fargate.html
	ecsContainerMetadataV4Endpoint, exists := os.LookupEnv("ECS_CONTAINER_METADATA_URI_V4")
	if !exists {
		log.Fatal("Could not get environment variable for the ECS Container Metadata URI! Exiting...")
	}

	// Load the Shared AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create API clients for ECS and Cloud Map services
	ecsClient := ecs.NewFromConfig(cfg)
	serviceDiscoveryClient := servicediscovery.NewFromConfig(cfg)

	// Fetch task metadata from the V4 Metadata Endpoint - here we get the task ARN
	taskMetadataFromEndpoint := services.GetTaskV4Metadata(ecsContainerMetadataV4Endpoint)

	// Fetch task information from the ECS API using the task ARN - here we want to get the task's service.
	taskInfoFromApi := services.DescribeTask(*ecsClient, taskMetadataFromEndpoint)

	// Get the task's service name from the "group" parameter
	serviceName := services.GetECSServiceForTask(taskInfoFromApi)

	// Fetch the service's service connect resources. We're interested in the Cloud Map Service ID (from DiscoveryArn)
	svcConnectResponse, enabled := services.GetServiceConnectResources(ecsClient, taskMetadataFromEndpoint.Cluster, serviceName)

	// If we get here, the service has Service Connect enabled, otherwise the above line will exit the program.
	if enabled {
		log.Printf("Discovery ARN for service %s - %s", serviceName, *svcConnectResponse.DiscoveryArn)
		log.Printf("Discovery name for service %s - %s", serviceName, *svcConnectResponse.DiscoveryName)

		for {

			// Here, we want to check the status of the task periodically after 5 seconds
			time.Sleep(time.Duration(time.Second * 5))

			// Fetch the task's health status from the ECS API
			taskInfoFromEcsApi := services.DescribeTask(*ecsClient, taskMetadataFromEndpoint)
			healthStatus := services.GetTaskHealthStatus(taskInfoFromEcsApi)
			lastKnownStatus := services.GetTaskLastKnownStatus(taskInfoFromEcsApi)

			log.Printf("Task %s status is %s and %s", *taskInfoFromApi.TaskArn, healthStatus, lastKnownStatus)

			// If the task's status is anything other than HEALTHY or the lifecycle state is not RUNNING and is either DEPROVISIONING,
			// DEPROVISIONING or STOPPING, we'd rather have the task de-registered from Cloud Map
			// than risk failed requests to the essential task since Cloud Map will route traffic regardless
			// of the instance's (task's) health status, if no health check is configured.
			// https://docs.aws.amazon.com/cloud-map/latest/dg/services-health-checks.html

			if healthStatus != "HEALTHY" || slices.Contains(ecsTaskShutdownStates, lastKnownStatus) {

				log.Printf("Attempting to de-register task's Cloud Map instance from the %s discovery name...", *svcConnectResponse.DiscoveryName)

				taskId := services.GetResourcePhysicalIdFromArn(*taskInfoFromEcsApi.TaskArn)
				cloudMapServiceId := services.GetResourcePhysicalIdFromArn(*svcConnectResponse.DiscoveryArn)

				deregistered := services.DeregisterTaskFromCloudMapService(*serviceDiscoveryClient, taskId, cloudMapServiceId)
				if deregistered {
					break
				}
			}
		}

		// If we get here, our task was de-registered from Cloud Map,
		// which means the task will be transitioning to STOPPED any time now.
		log.Println("I have done my job. Goodbye!")
	}
}
