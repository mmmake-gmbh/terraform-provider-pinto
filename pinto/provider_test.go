package pinto

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gitlab.com/whizus/gopinto"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_HasNeededResources(t *testing.T) {
	expectedResources := []string{
		"pinto_dns_zone",
		"pinto_dns_record",
	}

	resources := Provider(nil).ResourcesMap
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

	provider := Provider(nil)
	datasources := provider.DataSourcesMap
	require.Equal(t, len(expectedDatasources), len(datasources), "There are an unexpected number of registered datasource")

	for _, resource := range expectedDatasources {
		require.Contains(t, datasources, resource, "An expected datasource was not registered")
		require.NotNil(t, datasources[resource], "A datasource cannot have a nil schema")
	}
}


func TestProvider_ClientOverrideClientNil(t *testing.T) {
	provider := NewProvider(
		nil,
		"asdasd",
		"asdasd",
		"asdasd",
		)

	require.Nil(t, provider.client)
}

func TestProvider_ClientOverrideClientNotNil(t *testing.T) {
	provider := NewProvider(
		(*gopinto.APIClient)(NewMockClient(nil,nil)),
		"asdasd",
		"asdasd",
		"asdasd",
	)

	require.NotNil(t, provider.client)
}

func TestProviderConfiguration(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         providerFactory,
			PreventPostDestroyRefresh: false,
			Steps: []resource.TestStep{
				{
					Config: ProviderCfg,
					Destroy: false,
				},
			},
		},
	)
}
