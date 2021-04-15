---
page_title: "Resource pinto_dns_zone"
subcategory: "dns"
description: |-
  The pinto_dns_zone resource allows for the creation of DNS zones.
---

# Resource `pinto_dns_zone`

The pinto_dns_zone resource allows for the creation of DNS zones.

## Example Usage

```terraform
resource "pinto_dns_zone" "example_zone" {
  pinto_provider = "digitalocean"
  name           = "env0.co."
}
```

## Argument Reference

- `pinto_provider` - (String, Optional) Provider that pinto will use to store DNS entries (Required if provider is not set globally for the terraform provider)
- `pinto_environment` - (String, Optional) Environment at the provider that will be used to sore DNS entries
- `name` - (String, Required) The name of the zone

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the DNS zone
