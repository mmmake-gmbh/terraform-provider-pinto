package pinto

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProvider_HasNeededResources(t *testing.T) {
	expectedResources := []string{
		"pinto_dns_zone",
		"pinto_dns_record",
	}

	resources := Provider().ResourcesMap
	require.Equal(t, len(expectedResources), len(resources), "There are an unexpected number of registered resources")

	for _, resource := range expectedResources {
		require.Contains(t, resources, resource, "An expected resource was not registered")
		require.NotNil(t, resources[resource], "A resource cannot have a nil schema")
	}
}

func TestProvider_HasNeededDatasources(t *testing.T) {
	expectedDatasources := []string{
		"pinto_dns_zone",
		"pinto_dns_zones",
		"pinto_dns_record",
		"pinto_dns_records",
	}

	datasources := Provider().DataSourcesMap
	require.Equal(t, len(expectedDatasources), len(datasources), "There are an unexpected number of registered datasource")

	for _, resource := range expectedDatasources {
		require.Contains(t, datasources, resource, "An expected datasource was not registered")
		require.NotNil(t, datasources[resource], "A datasource cannot have a nil schema")
	}
}