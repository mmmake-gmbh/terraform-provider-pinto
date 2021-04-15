---
page_title: "Data Source pinto_dns_zone"
subcategory: "dns"
description: |-
  The zone data source allows for the retrieval of an existing zone.
---

# Data Source `pinto_dns_zone`

The zone data source allows for the retrieval of an existing zone.

## Example Usage

```terraform
data "pinto_dns_zone" "my_zone" {
  pinto_provider = "digitalocean"
  name           = "my.zone.com."
}
```

## Argument Reference

- `pinto_provider` - (String, Optional) Provider that pinto will use to store DNS entries (Required if provider is not set globally for the terraform provider)
- `pinto_environment` - (String, Optional) Environment at the provider that will be used to sore DNS entries
- `name` - (String, Required) The name of the zone 

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the zone
