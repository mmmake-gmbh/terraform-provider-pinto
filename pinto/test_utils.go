package pinto

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"os"
	"testing"

	"gitlab.com/whizus/gopinto"
)

type mockClient gopinto.APIClient
func NewMockClient(recordsApi gopinto.RecordsApi, zonesApi gopinto.ZonesApi) *mockClient {
	return &mockClient{
		RecordsApi: recordsApi,
		ZonesApi:   zonesApi,
	}
}

var ProviderCfg = `
provider "pinto" {
  base_url       = "https://pinto.irgendwo.co"
  token_url      = "https://auth.pinto.irgendwo.co/connect/token"
  client_id      = "machineclient"
  client_secret  = "Secret123$"
  client_scope   = "openapigateway,nexuscommon"
  pinto_provider = "digitalocean"
}`

// PreCheck TODO: remove this, the base_url will not be mandatory anymore in the future and will be set in the provider
func PreCheck(*testing.T) {
}
type service struct {
	client *mockClient
}

type mockRecordsApiService service

func (m mockRecordsApiService) ApiDnsRecordsDelete(ctx context.Context) gopinto.ApiApiDnsRecordsDeleteRequest {
	panic("implement me")
}

func (m mockRecordsApiService) ApiDnsRecordsDeleteExecute(r gopinto.ApiApiDnsRecordsDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockRecordsApiService) ApiDnsRecordsGet(ctx context.Context) gopinto.ApiApiDnsRecordsGetRequest {
	return gopinto.ApiApiDnsRecordsGetRequest{
		ApiService: m,
	}
}

func (m mockRecordsApiService) ApiDnsRecordsGetExecute(r gopinto.ApiApiDnsRecordsGetRequest) ([]gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	records := []gopinto.Record{}
	return records, &http.Response{StatusCode: 200}, gopinto.GenericOpenAPIError{}

}

func (m mockRecordsApiService) ApiDnsRecordsPost(ctx context.Context) gopinto.ApiApiDnsRecordsPostRequest {
	panic("implement me")
}

func (m mockRecordsApiService) ApiDnsRecordsPostExecute(r gopinto.ApiApiDnsRecordsPostRequest) (gopinto.Record, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

type mockZonesApiService service

func (m mockZonesApiService) ApiDnsZonesGet(ctx context.Context) gopinto.ApiApiDnsZonesGetRequest {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesGetExecute(r gopinto.ApiApiDnsZonesGetRequest) ([]gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesPost(ctx context.Context) gopinto.ApiApiDnsZonesPostRequest {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesPostExecute(r gopinto.ApiApiDnsZonesPostRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesZoneDelete(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneDeleteRequest {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesZoneDeleteExecute(r gopinto.ApiApiDnsZonesZoneDeleteRequest) (*http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesZoneGet(ctx context.Context, zone string) gopinto.ApiApiDnsZonesZoneGetRequest {
	panic("implement me")
}

func (m mockZonesApiService) ApiDnsZonesZoneGetExecute(r gopinto.ApiApiDnsZonesZoneGetRequest) (gopinto.Zone, *http.Response, gopinto.GenericOpenAPIError) {
	panic("implement me")
}

// define the providers used for the acceptance test cases
var providerFactory = map[string]func() (*schema.Provider, error){
	"pinto": func() (*schema.Provider, error) {
		acctest := os.Getenv("TF_ACC_MOCK")
		var provider *schema.Provider

		if acctest == "prod" {
			os.Setenv("PINTO_BASE_URL", "https://pinto.irgendwo.co")
			os.Setenv("PINTO_TOKEN_URL", "https://auth.pinto.irgendwo.co/connect/token")
			os.Setenv("PINTO_CLIENT_ID", "machineclient")
			os.Setenv("PINTO_CLIENT_SECRET", "Secret123$")
			os.Setenv("PINTO_CLIENT_SCOPE", "openapigateway,nexus")
			os.Setenv("PINTO_PROVIDER", "digitaloceano")
			os.Setenv("PINTO_ENVIRONMENT", "prod1")

			provider = Provider(nil)

			os.Unsetenv("PINTO_BASE_URL")
			os.Unsetenv("PINTO_TOKEN_URL")
			os.Unsetenv("PINTO_CLIENT_ID")
			os.Unsetenv("PINTO_CLIENT_SECRET")
			os.Unsetenv("PINTO_CLIENT_SCOPE")
			os.Unsetenv("PINTO_PROVIDER")
			os.Unsetenv("PINTO_ENVIRONMENT")

			return provider, nil
		}

		// TODO: Remove this when this in not required anymore
		os.Setenv("PINTO_BASE_URL", "https://mock.co")

		provider = Provider((*gopinto.APIClient)(NewMockClient(
			mockRecordsApiService{},
			mockZonesApiService{},
			)))

		os.Unsetenv("PINTO_ENVIRONMENT")

		return provider, nil
	},
}
