package pinto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProviderPintoDnsCreateZoneResource(t *testing.T) {
	name := "test_zone"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(createResources),
			Steps: []resource.TestStep{
				{
					Config: testAccConfigDNSZone(name),
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
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(changeRequest),
			Steps: []resource.TestStep{
				{
					Config: testAccConfigDNSZoneChange("env1.co"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.env0", "name", "env1.co"),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

func TestProviderPintoDnsImportZones(t *testing.T) {
	name := "test_zone"

	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(defaultMock),
			Steps: []resource.TestStep{
				{
					Config: testAccConfigDNSZone(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_zone.test_zone", "name", name+"."),
					),
					ExpectNonEmptyPlan: true,
					Destroy:            false,
				},
				{
					ResourceName:      `pinto_dns_zone.test_zone`,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func testAccConfigDNSZone(name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_zone" "%s" {
  	name              = "%s."
}`,
		name,
		name,
	)
}
func testAccConfigDNSZoneChange(name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_zone" "env0" {
  	name              = "%s"
}`, name)
}
