package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func zoneEntry(
	prefix string,
	name string,
	provider string,
	environment string,
) string {
	s := `
resource "pinto_dns_zone" "record_%s" {
	name 			  	= "%s"
	pinto_provider		= "%s"
	pinto_environment	= "%s"
}
`
	return fmt.Sprintf(
		s,
		prefix,
		name,
		provider,
		environment,
	)
}

func generateZoneTestEntry(prefix string) resource.TestStep {
	switch prefix {
	case "prod_digitalocean":
		return resource.TestStep{
			ResourceName: "pinto_dns_zone" + "." + prefix,
			PreConfig:    nil,
			Taint:        nil,
			Config: zoneEntry(
				prefix,
				"pinto",
				"prod1",
				"digitalocean",
			),
			Check:                   nil,
			Destroy:                 true,
			ImportState:             false,
			ImportStateId:           "",
			ImportStateIdPrefix:     "",
			ImportStateIdFunc:       nil,
			ImportStateCheck:        nil,
			ImportStateVerify:       false,
			ImportStateVerifyIgnore: nil,
		}
	}
	panic("please select a proper test case")
}

func TestResourceZoneEntry(t *testing.T) {
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
			Steps: []resource.TestStep{
				generateZoneTestEntry("prod_digitalocean"),
			},
		},
	)

}
