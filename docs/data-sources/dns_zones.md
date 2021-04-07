---
page_title: "Data Source dns_zones"
subcategory: "dns"
description: |-
  The zones data source allows for retrieval of existing zones.
---

# Data Source `dns_zones`

The zones data source allows for retrieval of existing zones.

## Example Usage

```terraform
data "pinto_dns_zone" "my_zones" {}
```

## Argument Reference

## Attributes Reference

- `id` - (String) The ID of the retrieved zones list
- `zones` - (Zones) A list of zones available at this provider, environment combination. See [Zones](#zones) for details.

### Zones

- `id` - (String) The ID of the zone
- `name` - (String) The name of the zone
