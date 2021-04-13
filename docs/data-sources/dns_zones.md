---
page_title: "Data Source pinto_dns_zones"
subcategory: "dns"
description: |-
  The pinto_dns_zones data source allows for the retrieval of existing zones.
---

# Data Source `pinto_dns_zones`

The pinto_dns_zones data source allows for the retrieval of existing zones.

## Example Usage

```terraform
data "pinto_dns_zone" "my_zones" {
  pinto_provider = "digitalocean"
}
```

## Argument Reference

- `pinto_provider` - (String, Optional) Provider that pinto will use to store DNS entries (Required if provider is not set globally for the terraform provider)
- `pinto_environment` - (String, Optional) Environment at the provider that will be used to sore DNS entries

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the retrieved zones list
- `zones` - (Zones) A list of zones available at this provider, environment combination. See [Zones](#zones) for details.

### Zones

- `id` - (String) The ID of the zone
- `name` - (String) The name of the zone
