package pinto

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// generate terraform definitions for test cases
func testResourceDNSRecord(
	prefix string,
	name string,
	_type string,
	class string,
	data string,
	zone string,
	environment string,
	provider string,
	ttl int,
) string {
	return fmt.Sprintf(`
resource "pinto_dns_record" "record_%s" {
	name 			  	= "%s"
	type 			  	= "%s"
	class 			  	= "%s"
	data 			  	= "%s"
	zone 			  	= "%s"
	pinto_environment  	= "%s"
	pinto_provider  	= "%s"
	ttl  				= %d
}
`,
		prefix,
		name,
		_type,
		class,
		data,
		zone,
		environment,
		provider,
		ttl,
	)
}

func TestResourceDnsRecord(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         providerFactory,
			Steps: []resource.TestStep{
				{
					Config: testResourceDNSRecord(
						"prod_env",
						"pinto",
						"A",
						"",
						"",
						"env0.co.",
						"prod1",
						"digitalocean",
						1800,
					),
					Destroy:                 true,
				},
			},
		},
	)
}
