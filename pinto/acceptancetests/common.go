package acceptancetests

import (
	"gitlab.com/whizus/terraform-provider-pinto/pinto"
	"os"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

// always use the same provider
var provider = pinto.Provider()

// TODO: remove this, the base_url will not be mandatory anymore in the future and will be set in the provider
// This can be removed as soon as the PINTO_BASE_URL is not required anymore.
// Having it optional is an intermediate requirement as it is not yet dinally defined.
func PreCheck(*testing.T) {
	os.Setenv("PINTO_BASE_URL", "https://baseurl.com")
}

// define the providers used for the acceptance test cases
var providerFactory = map[string]func() (*schema.Provider, error) {
	"pinto": func() (*schema.Provider, error) {
		return provider, nil
	},
}
