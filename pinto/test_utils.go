package pinto

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// always use the same provider
var provider = Provider(nil)

// TODO: Get mock environment
// default provider config
var ProviderCfg = `
provider "pinto" {
  base_url       = "https://pinto.irgendwo.co"
  token_url      = "https://auth.pinto.irgendwo.co/connect/token"
  client_id      = "machineclient"
  client_secret  = "Secret123$"
  client_scope   = "openapigateway,nexus"
  pinto_provider = "digitalocean"
}`

// TODO: remove this, the base_url will not be mandatory anymore in the future and will be set in the provider
// This can be removed as soon as the PINTO_BASE_URL is not required anymore.
// Having it optional is an intermediate requirement as it is not yet dinally defined.
func PreCheck(*testing.T) {
	os.Setenv("PINTO_BASE_URL", "https://baseurl.com")
	// os.Setenv("PINTO_TOKEN_URL", "https://auth.pinto.irgendwo.co/connect/token")
	// os.Setenv("PINTO_CLIENT_ID", "machineclient")
	// os.Setenv("PINTO_CLIENT_SECRET", "Secret123$")
	// os.Setenv("PINTO_CLIENT_SCOPE", "openapigateway,nexus")
	// os.Setenv("PINTO_PROVIDER", "digitalocean")
	// os.Setenv("PINTO_ENVIRONMENT", "prod1")
}

// define the providers used for the acceptance test cases
var providerFactory = map[string]func() (*schema.Provider, error){
	"pinto": func() (*schema.Provider, error) {
		return provider, nil
	},
}
