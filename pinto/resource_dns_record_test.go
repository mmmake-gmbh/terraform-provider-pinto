package pinto

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestProviderPintoDnsRecords(t *testing.T) {
	name := "record"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(defaultMock),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigResourceDNSRecord(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "name", name),
					),
				},
			},
		},
	)
}

func TestProviderPintoDnsCreateRecordResources(t *testing.T) {
	name := "record"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(createResources),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigResourceDNSRecord(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "name", name),
					),
					ExpectNonEmptyPlan: true,
				},
				resource.TestStep{
					Config: testAccConfigResourceDNSRecordWithoutTtl(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "name", name),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

func TestProviderPintoDnsChangeRecordResources(t *testing.T) {
	name := "record"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(changeRequest),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigResourceDNSRecord(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "data", "127.0.0.1"),
					),
				},
				resource.TestStep{
					Config: testAccConfigResourceChangeRecord(name + "_changed"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "data", "172.0.0.1"),
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "class", "FOO"),
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "ttl", "3600"),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

func TestProviderPintoDnsImportRecord(t *testing.T) {
	name := "record"
	resource.Test(
		t,
		resource.TestCase{
			IsUnitTest:        false,
			ProviderFactories: selectProviderConfiguration(defaultMock),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: testAccConfigResourceDNSRecord(name),
					Check: resource.ComposeTestCheckFunc(
						func(state *terraform.State) error {
							sr := state.RootModule().String()
							fmt.Sprintf("%s", sr)
							return nil
						},
						resource.TestCheckResourceAttr("pinto_dns_record.env0", "name", name),
					),
				},
				resource.TestStep{
					// TODO: Clarify the resource name
					ResourceName:      `pinto_dns_record.env0`,
					ImportState:       true,
					ImportStateVerify: true,
					ExpectError:       regexp.MustCompile("Error: invalid Import. ID has to be of format \"{type}/{name}/{zone}/{environment}/{provider}\""),
				},
			},
		},
	)
}

func testAccConfigResourceDNSRecord(name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_record" "env0" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
	zone              = "env0.co."
	name              = "%s"
	type              = "TXT"
	class             = "IN"
	data              = "127.0.0.1"
	ttl               = 1800
}
`, name)
}

// TODO: The class is required but it shouldn't be according to the spec
func testAccConfigResourceMissingClass(name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_record" "env0" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
	zone              = "env1.co."
	name              = "%s"
	class             = "IN"
	type              = "A"
	data              = "127.0.0.1"
}
`, name)
}

func testAccConfigResourceChangeRecord(name string) string {
	return fmt.Sprintf(`
data "pinto_dns_record" "%s" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
	zone              = "env0.co."
	name              = "record"
	type              = "A"
}

data "pinto_dns_records" "records" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
  	zone              = "env0.co."
}

resource "pinto_dns_record" "env0" {
	pinto_provider    = "digitalocean"
	pinto_environment = "prod1"
	zone              = "env1.co."
	name              = "%s"
	class             = "FOO"
	type              = "BAR"
	data              = "172.0.0.1"
  	ttl               = 3600
}
`, name, name)
}


func testAccConfigResourceDNSRecordWithoutTtl(name string) string {
	return fmt.Sprintf(`
resource "pinto_dns_record" "env0" {
  	pinto_provider    = "digitalocean"
  	pinto_environment = "prod1"
  	zone              = "env0.co."
  	name              = "%s"
  	type              = "TXT"
  	class             = "IN"
  	data              = "127.0.0.1"
}
`, name)
}