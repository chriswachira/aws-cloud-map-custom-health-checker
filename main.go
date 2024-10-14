package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/chriswachira/aws-cloud-map-custom-health-checker/services"
)

func main() {

	log.Println("Initializing AWS Cloud Map Custom Health Checker for Amazon ECS...")

	ECS_CONTAINER_METADATA_URI_V4, exists := os.LookupEnv("ECS_CONTAINER_METADATA_URI_V4")
	if !exists {
		log.Fatal("Could not get environment variable for the ECS Container Metadata URI! Exiting...")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	ecsClient := ecs.NewFromConfig(cfg)
	serviceDiscoveryClient := servicediscovery.NewFromConfig(cfg)

	// // services.GetECSServiceForTask(cfg, taskV4Metadata.TaskARN, taskV4Metadata.Cluster)
	// cluster := "arn:aws:ecs:eu-west-3:767397657267:cluster/RandyMarsh"
	// service := "pihole-dns-fargate"
	taskMetadataFromEndpoint := services.GetTaskV4Metadata(ECS_CONTAINER_METADATA_URI_V4)
	taskInfoFromApi := services.DescribeTask(*ecsClient, taskMetadataFromEndpoint)
	serviceName := services.GetECSServiceForTask(taskInfoFromApi)
	svcConnectResponse, enabled := services.GetServiceConnectResources(ecsClient, taskMetadataFromEndpoint.Cluster, serviceName)

	if enabled {
		log.Printf("Discovery ARN for service %s - %s", serviceName, *svcConnectResponse.DiscoveryArn)
		log.Printf("Discovery name for service %s - %s", serviceName, *svcConnectResponse.DiscoveryName)

		for {
			time.Sleep(time.Duration(time.Second * 5))
			healthStatus := services.GetTaskHealthStatus(taskInfoFromApi)

			log.Printf("Task %s status is %s", *taskInfoFromApi.TaskArn, healthStatus)

			if healthStatus != "HEALTHY" {
				log.Printf("Attempting to de-register task's Cloud Map instance from the %s discovery name...", *svcConnectResponse.DiscoveryName)

				deregistered := services.DeregisterTaskFromCloudMapService(*serviceDiscoveryClient, *taskInfoFromApi.TaskArn, *svcConnectResponse.DiscoveryArn)

				if deregistered {
					break
				}
			}
		}

		// If we get here that means our task was de-registered from Cloud Map,
		// which means the task will be transitioning to STOPPED any time now.
		log.Println("I have done my job. Goodbye!")
	}
}
