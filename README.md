# AWS Cloud Map Custom Health Checker

### Introduction
A lot of ECS customers use the Service Connect feature for **privately** inter-connecting their ECS services, usually their microservices.

When configuring Service Connect for your ECS service, Service Connect will automatically create a Cloud Map service for you if you don't select an existing one. You currently cannot configure a health check for the Cloud Map service when configuring Service Connect.

And when you don't configure a health check during a Cloud Map service creation, AWS Cloud Map will route traffic to service instances (EC2 instances, ECS tasks, EKS pods etc.) whether they are healthy or not.

This causes an issue where traffic from a Service Connect proxy is forwarded to tasks that are in unhealthy states, resulting in failed requests.

### Why not create a Cloud Map service manually and configure a health check?

Yes, you can create a Cloud Map service outside of ECS and use it to configure Service Connect for your ECS service.

The problem, however is that you cannot configure a Route 53 health check for a Cloud Map service created in a **private** DNS namespace.

### Okay, now what?
From the [official documentation](https://docs.aws.amazon.com/cloud-map/latest/dg/services-health-checks.html#services-health-check-custom), AWS Cloud Map recommends configuring custom health checks with a third-party health-checker for your Cloud Map service.

They suggest that the health-checker should update the Cloud Map service with a call to the `UpdateInstanceCustomHealthStatus` API and update the instance's status to `UNHEALTHY` in order for Cloud Map to stop routing traffic to that instance.

### The custom health-checker for a Cloud Map service...

This application tries to solve the above issues by periodically checking the health of your ECS tasks and de-registering the tasks from Cloud Map if the health check fails.

This is a more bruteforce way of telling Cloud Map to stop routing traffic to an ECS task that is not healthy. Since you cannot configure a Route 53 health check for private Cloud Map namespaces, this application does the health check for you and updates Cloud Map on the same.

### Contributing

I am a Golang newbie, any suggestions and recommendations to improve this program will be highly appreciated!
