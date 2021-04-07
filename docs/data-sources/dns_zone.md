---
page_title: "Data Source dns_zone"
subcategory: "dns"
description: |-
  The zone data source allows for retrieval of an existing zone.
---

# Data Source `dns_zone`

The zone data source allows for retrieval of an existing zone.

## Example Usage

```terraform
data "pinto_dns_zone" "my_zone" {
  name = "my.zone.com."
}
```

## Argument Reference

- `name` - (String, Required) The name of the zone 

## Attributes Reference

- `id` - (String) The ID of the zone
