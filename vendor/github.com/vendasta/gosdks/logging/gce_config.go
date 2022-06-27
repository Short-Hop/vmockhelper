package logging

import (
	"os"

	gce_metadata "cloud.google.com/go/compute/metadata"
)

type gkeCloudLoggingStrategy struct{}

// Compile time interface check
var _ CloudLoggingStrategy = &gkeCloudLoggingStrategy{}

func (g gkeCloudLoggingStrategy) GetMetaDataLabels(namespace, resourceName, appName string) (map[string]string, error) {
	projectID, err := gce_metadata.ProjectID()
	if err != nil {
		return nil, err
	}
	zone, err := gce_metadata.Zone()
	if err != nil {
		return nil, err
	}
	clusterName, err := gce_metadata.InstanceAttributeValue("cluster-name")
	if err != nil {
		return nil, err
	}

	nodeName, _ := gce_metadata.InstanceAttributeValue("node-name")
	if nodeName == "" {
		nodeName = os.Getenv("GKE_NODENAME")
	}

	dockerImage := os.Getenv("DOCKER_IMAGE")

	return map[string]string{
		"cluster_name":                     clusterName,
		"container_name":                   appName,
		"metadata.system_labels.node_name": nodeName,
		"namespace_name":                   namespace,
		"pod_name":                         resourceName,
		"docker_image":                     dockerImage,
		"project_id":                       projectID,
		// zone refers to the location of this container or instance. location refers to the location of the cluster master node.
		// We use zone to replace location as location is hard to get and it is not useful for us.
		"location": zone,
	}, nil
}
func (g gkeCloudLoggingStrategy) GetResourceType() string {
	return "k8s_container"
}

type cloudFunctionCloudLoggingStrategy struct{}

// Compile time interface check
var _ CloudLoggingStrategy = &cloudFunctionCloudLoggingStrategy{}

func (g cloudFunctionCloudLoggingStrategy) GetMetaDataLabels(namespace, resourceName, appName string) (map[string]string, error) {
	projectID, err := gce_metadata.ProjectID()
	if err != nil {
		return nil, err
	}
	cloudFunctionName := os.Getenv("K_SERVICE")
	cloudFunctionVersion := os.Getenv("K_REVISION")

	return map[string]string{
		"namespace_name":   namespace,
		"project_id":       projectID,
		"function_name":    cloudFunctionName,
		"function_version": cloudFunctionVersion,
	}, nil
}
func (g cloudFunctionCloudLoggingStrategy) GetResourceType() string {
	return "cloud_function"
}
