package pinto

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gitlab.com/whizus/gopinto"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

type ClientMock string
const (
	defaultMock ClientMock = "pinto-mock"
	badRequest = "pinto-mock-bad-request"
	testSystem = "pinto-test"
)

// PreCheck TODO: remove this, the base_url will not be mandatory anymore in the future and will be set in the provider
func PreCheck(*testing.T) {
	acc := os.Getenv("TF_ACC")
	if acc == testSystem {
		log.Printf("TF_ACC set to %s running against test system", acc)
	}
}

// either use the specific mocked provider for resource test, or override this with the "TC_ACC" variable for tests against a real system
func selectProviderConfiguration(mock ClientMock) map[string]func() (*schema.Provider, error) {
	acc := os.Getenv("TF_ACC")

	// use the build in client running against the default test-system
	if acc == testSystem {
		return map[string]func() (*schema.Provider, error){
			"pinto": providerConfigurations[testSystem],
		}
	}

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
	badRequest: func() (*schema.Provider, error) {
		var provider *schema.Provider
		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockBadRequestRecordsApiService{},
			mockBadRequestZonesApiService{},
		)))

		return provider, nil
	},
	testSystem: func() (*schema.Provider, error) {
		var provider *schema.Provider

		os.Setenv("PINTO_BASE_URL", "https://pinto.irgendwo.co")
		os.Setenv("PINTO_TOKEN_URL", "https://auth.pinto.irgendwo.co/connect/token")
		os.Setenv("PINTO_CLIENT_ID", "machineclient")
		os.Setenv("PINTO_CLIENT_SECRET", "Secret123$")
		os.Setenv("PINTO_CLIENT_SCOPE", "openapigateway,nexus")
		os.Setenv("PINTO_PROVIDER", "digitaloceano")
		os.Setenv("PINTO_ENVIRONMENT", "prod1")

		provider = Provider(nil)

		return provider, nil
	},
}

// unit tests
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
		"mock-api-key",
		"mock-provider",
		"mock-environment",
	)

	require.Nil(t, provider.client)
}

func TestProvider_ClientOverrideClientNotNil(t *testing.T) {
	provider := NewProvider(
		(*gopinto.APIClient)(NewMockClient(nil, nil)),
		"mock-api-key",
		"mock-provider",
		"mock-environment",
	)

	require.NotNil(t, provider.client)
}

// acctests
func TestProviderUnknownAttributes(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         selectProviderConfiguration(badRequest),
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
var providerExampleSteps = []resource.TestStep{
		resource.TestStep{
		Config: `
data "pinto_dns_zone" "zone1" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  name              = "env0.co."
}

data "pinto_dns_zones" "zones" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
}

data "pinto_dns_record" "record" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  zone              = "env0.co."
  name              = "pinto"
  type              = "A"
}

data "pinto_dns_records" "records" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
  	zone              = "env0.co."
}
			`,
		Check: func(state *terraform.State) error {
			resourceName := "data.pinto_dns_records.records"
			_, ok := state.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("resource %s not found in state", resourceName)
			}

			resourceName = "data.pinto_dns_zone.zone1"
			_, ok = state.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("resource %s not found in state", "")
			}

			resourceName = "data.pinto_dns_zones.zones"
			_, ok = state.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("resource %s not found in state", resourceName)
			}

			return nil
		},
	},
	resource.TestStep{
		ResourceName: "env0.co.prod1.digitalocean.",
		Config: `
resource "pinto_dns_zone" "env0" {
  	pinto_environment = "prod1"
  	pinto_provider    = "digitalocean"
	name = "env0.co."
}
			`,
		Destroy: true,
	},

	resource.TestStep{
		ResourceName: "TXT/somewhere/env0.co./prod1/digitalocean",
		Config: `
resource "pinto_dns_record" "env0" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  zone              = "env0.co."
  name              = "somewhere"
  type              = "TXT"
  class             = "IN"
  data              = "127.0.0.1"
  ttl               = 1800
}
			`,
		Destroy: true,
	},
}

func TestProviderExampleStepsDefaultMock(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         selectProviderConfiguration(defaultMock),
			PreventPostDestroyRefresh: false,
			Steps: providerExampleSteps,
		},
	)
}

func TestProviderExampleStepsBadRequest(t *testing.T) {
	dataExampleStep := providerExampleSteps[0]
	dataExampleStep.ExpectError = regexp.MustCompile("Error: 400 Bad Request")

	resourceZoneExampleStep := providerExampleSteps[1]
	resourceZoneExampleStep.Config =  `
resource "pinto_dns_zone" "env0" {
  	pinto_environment = "prod1"
  	pinto_provider    = "digitalocean"
	name = "env0.co."
}
`
	resourceRecordExampleStep := providerExampleSteps[2]

	steps := []resource.TestStep{
		dataExampleStep,
		resourceZoneExampleStep,
		resourceRecordExampleStep,
	}

	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         selectProviderConfiguration(badRequest),
			Steps: steps,
		},
	)
}

type service struct {
	client *mockClient
}

type mockClient gopinto.APIClient

// default api mocks returning 20x results
type mockRecordsApiService service
type mockZonesApiService service

type mockBadRequestRecordsApiService service
type mockBadRequestZonesApiService service


func (m mockBadRequestZonesApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	return gopinto.ApiApiDnsZonesGetRequest{
		ApiService: m,
	}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	return gopinto.ApiApiDnsZonesPostRequest{
		ApiService: m,
	}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	return gopinto.ApiApiDnsZonesZoneDeleteRequest{
		ApiService: m,
	}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	return gopinto.ApiApiDnsZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockBadRequestZonesApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}


func (jj mockBadRequestRecordsApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	return gopinto.ApiApiDnsRecordsDeleteRequest{
		ApiService: jj,
	}
}

func (jj mockBadRequestRecordsApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (jj mockBadRequestRecordsApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: jj,
	}
}

func (jj mockBadRequestRecordsApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Record{}, &http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (jj mockBadRequestRecordsApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	panic("implement me")
}

func (jj mockBadRequestRecordsApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func NewMockClient(recordsApi gopinto.RecordsApi, zonesApi gopinto.ZonesApi) *mockClient {
	return &mockClient{
		RecordsApi: recordsApi,
		ZonesApi:   zonesApi,
	}
}

func (m mockRecordsApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	return gopinto.ApiApiDnsRecordsDeleteRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Record{
		{
			Name:  "somewhere",
			Type:  "A",
			Class: "IN",
			Ttl:   toInt64(1800),
			Data:  "127.0.0.1",
		},
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}

}

func (m mockRecordsApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	return gopinto.ApiApiDnsRecordsPostRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{
		Name:  "pinto",
		Type:  "A",
		Class: "IN",
		Ttl:   toInt64(1800),
		Data:  "127.0.0.1",
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}


func (m mockZonesApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	return gopinto.ApiApiDnsZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Zone{
		{
			Name: "env0.co.",
		},
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	return gopinto.ApiApiDnsZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{
		Name: "env0.co.",
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	return gopinto.ApiApiDnsZonesZoneDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	return gopinto.ApiApiDnsZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	zone := gopinto.Zone{
		Name: "env0.co.",
	}
	return zone, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func toInt64(x int64) *int64 {
	return &x
}
