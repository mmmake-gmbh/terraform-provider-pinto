package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)


// TODO: clarify records and record
// func dnsRecords(prefix string, environment string, zone string) string {
// 	s := `
// resource "pinto_dns_records" "record_%s" {
//   	pinto_environment = "%s"
//   	zone              = "%s"
// }
// `
// 	return fmt.Sprintf(s, prefix, environment, zone)
// }

// generate terraform definitions for test cases
func dnsRecord(
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
	s := `
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
`
	return fmt.Sprintf(
		s,
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

// TODO: extend this for different test cases
// used to generate different 'defined' test cases.
// panics on not resolvable testcases
func generateTestRecordEntry(prefix string) resource.TestStep {
	switch prefix {
	case "prod_env_zero":
		return resource.TestStep{
			ResourceName:              "pinto_dns_records" + "." + prefix,
			PreConfig:                 nil,
			Taint:                     nil,
			Config:                    dnsRecord(
				prefix,
				"pinto",
				"A",
				"",
				"",
				"env0.co.",
				"prod1",
				"digitalocean",
				1800,
			),
			Check:                     nil,
			Destroy:                   true,
			ImportState:               false,
			ImportStateId:             "",
			ImportStateIdPrefix:       "",
			ImportStateIdFunc:         nil,
			ImportStateCheck:          nil,
			ImportStateVerify:         false,
			ImportStateVerifyIgnore:   nil,
		}
		break
	}
	panic("please select a proper test case")
}

func TestResourceDnsRecord(t *testing.T) {
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:                false,
			PreCheck:                  func() { PreCheck(t) },
			ProviderFactories:         providerFactory,
			ProtoV5ProviderFactories:  nil,
			ExternalProviders:         nil,
			PreventPostDestroyRefresh: false,
			CheckDestroy:              nil,
			ErrorCheck:                nil,
			Steps:                     []resource.TestStep{
				generateTestRecordEntry("prod_env_zero"),
			},
		},
	)

}

