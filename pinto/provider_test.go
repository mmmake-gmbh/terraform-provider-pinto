package pinto

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/whizus/gopinto"

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
type mockZonesApiService service

type mockRecordsCreateApiService service

func (m mockRecordsCreateApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	return gopinto.ApiApiDnsRecordsDeleteRequest{
		ApiService: m,
	}
}
func (m mockRecordsCreateApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsCreateApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: m,
	}
}
func (m mockRecordsCreateApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Record{}, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsCreateApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	return gopinto.ApiApiDnsRecordsPostRequest{
		ApiService: m,
	}
}
func (m mockRecordsCreateApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{
		Name:  "record",
		Type:  "A",
		Ttl:   toInt32(1800),
		Class: "IN",
		Data:  "127.0.0.1",
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

type mockZonesCreateApiService service

func (m mockZonesCreateApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	return gopinto.ApiApiDnsZonesGetRequest{
		ApiService: m,
	}
}
func (m mockZonesCreateApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Zone{}, &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesCreateApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	return gopinto.ApiApiDnsZonesPostRequest{
		ApiService: m,
	}
}
func (m mockZonesCreateApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{
			Name: "env0.co.",
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesCreateApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	return gopinto.ApiApiDnsZonesZoneDeleteRequest{
		ApiService: m,
	}
}
func (m mockZonesCreateApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesCreateApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	return gopinto.ApiApiDnsZonesZoneGetRequest{
		ApiService: m,
	}
}
func (m mockZonesCreateApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{
			Name: "env0.co.",
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}

type mockRecordsBadApiService service

func (m mockRecordsBadApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	return gopinto.ApiApiDnsRecordsDeleteRequest{
		ApiService: m,
	}
}
func (m mockRecordsBadApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsBadApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: m,
	}
}
func (m mockRecordsBadApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Record{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsBadApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	return gopinto.ApiApiDnsRecordsPostRequest{
		ApiService: m,
	}
}
func (m mockRecordsBadApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{
		Name:  "record",
		Type:  "A",
		Class: "IN",
		Ttl:   toInt32(1800),
		Data:  "127.0.0.1",
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

type mockZonesBadApiService service

func (m mockZonesBadApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	return gopinto.ApiApiDnsZonesGetRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	return gopinto.ApiApiDnsZonesPostRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	return gopinto.ApiApiDnsZonesZoneDeleteRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

func (m mockZonesBadApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	return gopinto.ApiApiDnsZonesZoneGetRequest{
		ApiService: m,
	}
}

func (m mockZonesBadApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{}, &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Error: 400 Bad Request")),
	}, gopinto.GenericOpenAPIError{}
}

type mockZonesChangeApiService service

func (m mockZonesChangeApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	return gopinto.ApiApiDnsZonesGetRequest{
		ApiService: m,
	}
}
func (m mockZonesChangeApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Zone{
			{
				Name: "env1.co.",
			},
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesChangeApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	return gopinto.ApiApiDnsZonesPostRequest{
		ApiService: m,
	}
}
func (m mockZonesChangeApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{
			Name: "env0-changed.co.",
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesChangeApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	return gopinto.ApiApiDnsZonesZoneDeleteRequest{
		ApiService: m,
	}
}
func (m mockZonesChangeApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockZonesChangeApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	return gopinto.ApiApiDnsZonesZoneGetRequest{
		ApiService: m,
	}
}
func (m mockZonesChangeApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Zone{
			Name: "env0-changed.co.",
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}

type mockRecordsChangeApiService service

func (m mockRecordsChangeApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	return gopinto.ApiApiDnsRecordsDeleteRequest{
		ApiService: m,
	}
}
func (m mockRecordsChangeApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	return &http.Response{
		StatusCode: 200,
	}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsChangeApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: m,
	}
}
func (m mockRecordsChangeApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return []gopinto.Record{
			{
				Name:  "record",
				Type:  "A",
				Class: "IN",
				Ttl:   toInt32(1800),
			},
		}, &http.Response{
			StatusCode: 200,
		}, gopinto.GenericOpenAPIError{}
}
func (m mockRecordsChangeApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	return gopinto.ApiApiDnsRecordsPostRequest{
		ApiService: m,
	}
}
func (m mockRecordsChangeApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	return gopinto.Record{
		Name:  "record",
		Type:  "A",
		Class: "IN",
		Ttl:   toInt32(1800),
		Data:  "127.0.0.1",
	}, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
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
			Name:  "record",
			Type:  "A",
			Class: "IN",
			Ttl:   toInt32(1800),
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
		Ttl:   toInt32(1800),
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
			Name: "env0-1.co.",
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
		Name: "env0-2.co.",
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
		Name: "env0-1.co.",
	}
	return zone, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}
}

func toInt32(x int32) *int32 {
	return &x
}
