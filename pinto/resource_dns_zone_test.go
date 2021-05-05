package pinto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProviderPintoDnsCreateZoneResource(t *testing.T) {
	name := "test_zone"
	provider := "digitalocean"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(createResources),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigDNSZone(provider, name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.test_zone", "name", name+"."),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

func TestProviderPintoDnsChangeZoneResource(t *testing.T) {
	name := "env0.co."
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(changeRequest),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigDNSZoneChange(name, name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.env0", "name", name),
					),
					ExpectNonEmptyPlan: true,
				},
				resource.TestStep{
					Config: testAccConfigDNSZoneChange(name, "env1.co."),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.env0", "name", "env1.co."),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

func TestProviderPintoDnsImportZones(t *testing.T) {
	name := "test_zone"
	provider := "digitalocean"

	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(defaultMock),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigDNSZone(provider, name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.test_zone", "name", name+"."),
					),
					ExpectNonEmptyPlan: true,
					Destroy:            false,
				},
				resource.TestStep{
					ResourceName:      `pinto_dns_zone.test_zone`,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func testAccConfigDNSZone(provider, name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_zone" "%s" {
  	pinto_provider    = "%s"
  	pinto_environment = "%s"
  	name              = "%s."
}`,
	name,
	provider,
	"prod1",
	name,
	)
}
func testAccConfigDNSZoneChange(oldName string, newName string) string {
	return fmt.Sprintf(`
data "pinto_dns_zone" "env0" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  name              = "%s"
}

data "pinto_dns_zones" "zones" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
}

resource "pinto_dns_zone" "env0" {
  	pinto_provider    = "digitalocean"
  	pinto_environment = "prod1"
  	name              = "%s"
}`, oldName, newName)
}
