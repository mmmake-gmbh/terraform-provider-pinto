package pinto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataDnsRecord(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         providerFactory,
			PreventPostDestroyRefresh: false,
			Steps: []resource.TestStep{
				// TODO: Investigate, fails with
				// testing_new.go:49: no "id" found in attributes
				// testing_new.go:49: no "id" found in attributes
				// {
				// 	ResourceName:              "",
				// 	Config:                    testDataDNSRecord("dns_record", "mock", "mock", "mock.co", "entry", "A"),
				// },
			},
		},
	)
}

func testDataDNSRecord(prefix string, pinto_provider string, pinto_environment string, zone string, name string, _type string,) string {
	return fmt.Sprintf(`
data "pinto_dns_record" "record_%s" {
  pinto_provider    = "%s"
  pinto_environment = "%s"
  zone              = "%s"
  name              = "%s"
  type              = "%s"
}`,
		prefix,
		pinto_provider,
		pinto_environment,
		zone,
		name,
		_type,
	)
}
