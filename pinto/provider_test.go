package pinto

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"testing"

	gopinto "github.com/camaoag/project-pinto-sdk-go"
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

		os.Setenv("PINTO_BASE_URL", "https://mock.co")
		os.Setenv("PINTO_CREDENTIALS_ID", "4d4fe4ac-586e-4121-9603-43acf2b0ce8d")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsApiService{},
			mockZonesApiService{},
		)))

		return provider, nil
	},
	brokenApi: func() (*schema.Provider, error) {
		var provider *schema.Provider

		os.Setenv("PINTO_BASE_URL", "https://mock.co")
		os.Setenv("PINTO_CREDENTIALS_ID", "4d4fe4ac-586e-4121-9603-43acf2b0ce8d")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsBadApiService{},
			mockZonesBadApiService{},
		)))

		return provider, nil
	},
	changeRequest: func() (*schema.Provider, error) {
		var provider *schema.Provider

		os.Setenv("PINTO_BASE_URL", "https://mock.co")
		os.Setenv("PINTO_CREDENTIALS_ID", "4d4fe4ac-586e-4121-9603-43acf2b0ce8d")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsChangeApiService{},
			mockZonesChangeApiService{},
		)))

		return provider, nil
	},
	createResources: func() (*schema.Provider, error) {
		var provider *schema.Provider

		os.Setenv("PINTO_BASE_URL", "https://mock.co")
		os.Setenv("PINTO_CREDENTIALS_ID", "4d4fe4ac-586e-4121-9603-43acf2b0ce8d")

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
				{
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
				{
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
				{
					ExpectError: regexp.MustCompile("Error: 400 Bad Request"),
					Config: `
 data "pinto_dns_zones" "zones" {
 	pinto_provider    = "unknown"
 	pinto_environment = "unknown"
 }
 			`,
					Destroy: true,
				},
				{
					Config: `
 resource "pinto_dns_zone" "unknown" {
   	pinto_environment = "unknown"
   	pinto_provider    = "unknown"
 	name = "unknown.co."
 }
 			`,
					Destroy: true,
				},
				{
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

func newRecord(name string) gopinto.Record {
	return gopinto.Record{
		Name:  name,
		Type:  "TXT",
		Class: "IN",
		Data:  "127.0.0.1",
		Ttl:   toInt32(1800),
	}
}

func newZone(name string) gopinto.Zone {
	return gopinto.Zone{
		Name: name,
	}
}

type mockClient gopinto.APIClient

// default api mocks returning 20x results
type mockRecordsApiService service

func (m mockRecordsApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	return gopinto.ApiDnsApiRecordsDeleteRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	return gopinto.ApiDnsApiRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	records := []gopinto.Record{
		newRecord("testrecord"),
	}
	return records, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	return gopinto.ApiDnsApiRecordsPostRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return newRecord("testrecord"), &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

type mockZonesApiService service

func (m mockZonesApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	return gopinto.ApiDnsApiZonesDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	return gopinto.ApiDnsApiZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	zones := []gopinto.Zone{
		newZone("env0.co."),
	}
	return zones, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	return gopinto.ApiDnsApiZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env0.co."), &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	return gopinto.ApiDnsApiZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env0.co."), &http.Response{
		StatusCode: 201,
	}, gopinto.GenericOpenAPIError{}
}

type mockRecordsCreateApiService service

func (m mockRecordsCreateApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	return gopinto.ApiDnsApiRecordsDeleteRequest{
		ApiService: m,
	}
}

func (m mockRecordsCreateApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 201,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsCreateApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	return gopinto.ApiDnsApiRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsCreateApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	records := []gopinto.Record{
		newRecord("testrecord"),
	}
	return records, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsCreateApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	return gopinto.ApiDnsApiRecordsPostRequest{
		ApiService: m,
	}
}

func (m mockRecordsCreateApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{}, &http.Response{}, gopinto.GenericOpenAPIError{}
}

type mockZonesCreateApiService service

func (m mockZonesCreateApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	return gopinto.ApiDnsApiZonesDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesCreateApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesCreateApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	return gopinto.ApiDnsApiZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesCreateApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	zones := []gopinto.Zone{
		newZone("env0.co."),
	}
	return zones, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesCreateApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	return gopinto.ApiDnsApiZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesCreateApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env0.co."), &http.Response{
		StatusCode: 201,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesCreateApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	return gopinto.ApiDnsApiZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesCreateApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env0.co."), &http.Response{
		StatusCode: 201,
	}, gopinto.GenericOpenAPIError{}
}

type mockRecordsBadApiService service

func (m mockRecordsBadApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	return gopinto.ApiDnsApiRecordsDeleteRequest{
		ApiService: m,
	}
}

func (m mockRecordsBadApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsBadApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	return gopinto.ApiDnsApiRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsBadApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return nil, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsBadApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	return gopinto.ApiDnsApiRecordsPostRequest{
		ApiService: m,
	}
}

func (m mockRecordsBadApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

type mockZonesBadApiService service

func (m mockZonesBadApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	return gopinto.ApiDnsApiZonesDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	return gopinto.ApiDnsApiZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return nil, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	return gopinto.ApiDnsApiZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	return gopinto.ApiDnsApiZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

type mockZonesChangeApiService service

func (m mockZonesChangeApiService) DnsApiZonesDelete(ctx context.Context) gopinto.ApiDnsApiZonesDeleteRequest {
	return gopinto.ApiDnsApiZonesDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesChangeApiService) DnsApiZonesDeleteExecute(r gopinto.ApiDnsApiZonesDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesChangeApiService) DnsApiZonesGet(ctx context.Context) gopinto.ApiDnsApiZonesGetRequest {
	return gopinto.ApiDnsApiZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesChangeApiService) DnsApiZonesGetExecute(r gopinto.ApiDnsApiZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	zones := []gopinto.Zone{
		newZone("env1.co"),
	}
	return zones, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesChangeApiService) DnsApiZonesPost(ctx context.Context) gopinto.ApiDnsApiZonesPostRequest {
	return gopinto.ApiDnsApiZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesChangeApiService) DnsApiZonesPostExecute(r gopinto.ApiDnsApiZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env1.co."), &http.Response{}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesChangeApiService) DnsApiZonesZoneGet(ctx context.Context, zone string) gopinto.ApiDnsApiZonesZoneGetRequest {
	return gopinto.ApiDnsApiZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesChangeApiService) DnsApiZonesZoneGetExecute(r gopinto.ApiDnsApiZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return newZone("env1.co."), &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

type mockRecordsChangeApiService service

func (m mockRecordsChangeApiService) DnsApiRecordsDelete(ctx context.Context) gopinto.ApiDnsApiRecordsDeleteRequest {
	return gopinto.ApiDnsApiRecordsDeleteRequest{
		ApiService: m,
	}
}

func (m mockRecordsChangeApiService) DnsApiRecordsDeleteExecute(r gopinto.ApiDnsApiRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsChangeApiService) DnsApiRecordsGet(ctx context.Context) gopinto.ApiDnsApiRecordsGetRequest {
	return gopinto.ApiDnsApiRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsChangeApiService) DnsApiRecordsGetExecute(r gopinto.ApiDnsApiRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	records := []gopinto.Record{
		newRecord("testrecord2"),
	}
	return records, &http.Response{}, gopinto.GenericOpenAPIError{}
}

func (m mockRecordsChangeApiService) DnsApiRecordsPost(ctx context.Context) gopinto.ApiDnsApiRecordsPostRequest {
	return gopinto.ApiDnsApiRecordsPostRequest{
		ApiService: m,
	}
}

func (m mockRecordsChangeApiService) DnsApiRecordsPostExecute(r gopinto.ApiDnsApiRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return newRecord("testrecord2"), &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}

func toInt32(x int32) *int32 {
	return &x
}
