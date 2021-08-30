package pinto

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/camaoag/project-pinto-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/stretchr/testify/require"
)

type ClientMock string

const (
	defaultMock     ClientMock = "pinto-mock"
	changeRequest              = "pinto-mock-change-request"
	createResources            = "pinto-mock-create-resources"
	brokenApi                  = "pinto-mock-broken-api"
)

// either use the specific mocked provider for resource test, or override this with the "TC_ACC" variable for tests against a real system
func selectProviderConfiguration(mock ClientMock) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"pinto": providerConfigurations[mock],
	}

}

var providerConfigurations = map[ClientMock]func() (*schema.Provider, error){
	defaultMock: func() (*schema.Provider, error) {
		var provider *schema.Provider
		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsApiService{},
			mockZonesApiService{},
		)))

		return provider, nil
	},
	brokenApi: func() (*schema.Provider, error) {
		var provider *schema.Provider
		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsBadApiService{},
			mockZonesBadApiService{},
		)))

		return provider, nil
	},
	changeRequest: func() (*schema.Provider, error) {
		var provider *schema.Provider
		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsChangeApiService{},
			mockZonesChangeApiService{},
		)))

		return provider, nil
	},
	createResources: func() (*schema.Provider, error) {
		var provider *schema.Provider
		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsCreateApiService{},
			mockZonesCreateApiService{},
		)))

		return provider, nil
	},
}

func NewMockClient(recordsApi gopinto.RecordsApi, zonesApi gopinto.ZonesApi) *mockClient {
	return &mockClient{
		RecordsApi: recordsApi,
		ZonesApi:   zonesApi,
	}
}

// unit tests to validate that the provider implements the expected resources
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

// unit tests to validate that the provider implements the expected datasources
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

// handle errors whenever the client returns errors
// currently even a 500 is interpreted as 400 in the terraform provider,
// but with the same result "a http.statusCode somewhere above 400" aka an error
func TestProviderUnknownAttributes(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			ProviderFactories:         selectProviderConfiguration(brokenApi),
			PreventPostDestroyRefresh: false,
			Steps: []resource.TestStep{
				resource.TestStep{
					ExpectError: regexp.MustCompile("Error: 400 Bad Request"),
					Config: `
 data "pinto_dns_records" "records_unknown_provider" {
 	pinto_provider    = "unknown"
 	pinto_environment = "unknown"
 	zone              = "somewhere.co."
 }
 			`,
					Destroy: true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ibm_org.testacc_ds_org", "org", "organization"),
						resource.TestCheckResourceAttrSet("data.ibm_account.testacc_acc", "id"),
					),
				},
				resource.TestStep{
					ExpectError: regexp.MustCompile("Error: 400 Bad Request"),
					Config: `
 data "pinto_dns_zone" "zone1" {
   	pinto_provider    = "unkown"
   	pinto_environment = "unknown"
   	name              = "unknown"
 }
 			`,
					Destroy: true,
				},
				resource.TestStep{
					ExpectError: regexp.MustCompile("Error: 400 Bad Request"),
					Config: `
 data "pinto_dns_zones" "zones" {
 	pinto_provider    = "unknown"
 	pinto_environment = "unknown"
 }
 			`,
					Destroy: true,
				},
				resource.TestStep{
					Config: `
 resource "pinto_dns_zone" "unknown" {
   	pinto_environment = "unknown"
   	pinto_provider    = "unknown"
 	name = "unknown.co."
 }
 			`,
					Destroy: true,
				},
				resource.TestStep{
					Config: `
 resource "pinto_dns_record" "env0_unknown" {
   pinto_provider    = "unknown"
   pinto_environment = "unknown"
   zone              = "unknown.co."
   name              = "unknown"
   type              = "TXT"
   class             = "IN"
   data              = "127.0.0.1"
   ttl               = 1800
 }
 			`,
					Destroy: true,
				},
			},
		},
	)
}

type service struct {
	client *mockClient
}

type mockClient gopinto.APIClient

// default api mocks returning 20x results
type mockRecordsApiService service

func (m mockRecordsApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	panic("implement me")
}

func (m mockRecordsApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	panic("implement me")
}

func (m mockRecordsApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	panic("implement me")
}

func (m mockRecordsApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockZonesApiService service

func (m mockZonesApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	panic("implement me")
}

func (m mockZonesApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockRecordsCreateApiService service

func (m mockRecordsCreateApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	panic("implement me")
}

func (m mockRecordsCreateApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsCreateApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	panic("implement me")
}

func (m mockRecordsCreateApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsCreateApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	panic("implement me")
}

func (m mockRecordsCreateApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockZonesCreateApiService service

func (m mockZonesCreateApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	panic("implement me")
}

func (m mockZonesCreateApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockRecordsBadApiService service

func (m mockRecordsBadApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	panic("implement me")
}

func (m mockRecordsBadApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsBadApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	panic("implement me")
}

func (m mockRecordsBadApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsBadApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	panic("implement me")
}

func (m mockRecordsBadApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockZonesBadApiService service

func (m mockZonesBadApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	panic("implement me")
}

func (m mockZonesBadApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockZonesChangeApiService service

func (m mockZonesChangeApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	panic("implement me")
}

func (m mockZonesChangeApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockRecordsChangeApiService service

func (m mockRecordsChangeApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	panic("implement me")
}

func (m mockRecordsChangeApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsChangeApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	panic("implement me")
}

func (m mockRecordsChangeApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsChangeApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	panic("implement me")
}

func (m mockRecordsChangeApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}
