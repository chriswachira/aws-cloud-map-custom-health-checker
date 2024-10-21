package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/chriswachira/aws-cloud-map-custom-health-checker/services"
)

const (
	exitCodeSuccess      = 0
	exitCodeProgramError = 1
)

func main() {

	// A task can be healthy and ACTIVATING status at the same time, meaning that if
	// this app container, that is configured as a dependency for an essential container,
	// is launched after the essential becomes HEALTHY, this app will deregister the
	// task from its Cloud Map service before it transitions to the RUNNING state.
	log.Println("Waiting for task to fully initialize; sleeping for 60 seconds...")
	time.Sleep(time.Duration(time.Second * 60))

	// Initialize a channel that we can use to capture process signals from
	// the ECS Service Scheduler.
	log.Println("Initializing AWS Cloud Map Custom Health Checker for Amazon ECS...")
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

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

	// Get task information from V4 endpoint and ECS API
	taskMetadataFromEndpoint := services.GetTaskV4Metadata(ecsContainerMetadataV4Endpoint)
	// taskDefinitionDetails := services.GetTaskDefinitionDetails(*ecsClient, taskMetadataFromEndpoint.Family, taskMetadataFromEndpoint.Revision)
	taskInfoFromEcsApi := services.DescribeTask(*ecsClient, taskMetadataFromEndpoint)

	// Get the task's service name from the "group" parameter
	serviceName := services.GetECSServiceForTask(taskInfoFromEcsApi)

	// Fetch the service's service connect resources. We're interested in the Cloud Map Service ID (from DiscoveryArn)
	svcConnectResponse, enabled := services.GetServiceConnectResources(ecsClient, taskMetadataFromEndpoint.Cluster, serviceName)

	// If we get here, the service has Service Connect enabled, otherwise the above line will exit the program.
	if enabled {
		log.Printf("Discovery ARN for service %s - %s", serviceName, *svcConnectResponse.DiscoveryArn)
		log.Printf("Discovery name for service %s - %s", serviceName, *svcConnectResponse.DiscoveryName)
		log.Printf("Listening for an interrupt signal from the ECS Service Scheduler...")

		// Capture a process signal from the main go-routine. Here we want to capture the
		// SIGTERM signal that is sent from the ECS Service Scheduler when a user stops the task or
		// when the task becomes unhealthy. The SIGINT signal is mostly for local development.
		// We obviously can't catch the SIGKILL signal so it's not included in the signal capture below.
		sig := <-signalChannel
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Println("Received an interrupt signal from the ECS Service Scheduler...")
			log.Printf("Attempting to de-register task's Cloud Map instance from the %s discovery name...", *svcConnectResponse.DiscoveryName)

			taskId := services.GetResourcePhysicalIdFromArn(*taskInfoFromEcsApi.TaskArn)
			cloudMapServiceId := services.GetResourcePhysicalIdFromArn(*svcConnectResponse.DiscoveryArn)

			deregistered := services.DeregisterTaskFromCloudMapService(*serviceDiscoveryClient, taskId, cloudMapServiceId)
			if deregistered {

				// If we get here, our task was de-registered from Cloud Map,
				// which means the task will be transitioning to STOPPED any time now.
				log.Println("I have done my job. Goodbye!")
				os.Exit(exitCodeSuccess)
			}
		}
	}
}
