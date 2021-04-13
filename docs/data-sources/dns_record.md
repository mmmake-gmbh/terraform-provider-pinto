---
page_title: "Data Source pinto_dns_record"
subcategory: "dns"
description: |-
  The records data source allows for the retrieval of an existing DNS record.
---

# Data Source `pinto_dns_record`

 The records data source allows for the retrieval of an existing DNS record.

## Example Usage

```terraform
data "pinto_dns_record" "my_record" {
  pinto_providoer = "digitalocean"
  zone            = "my.zone.com."
  name            = "my_record_name"
  type            = "A"
}
```

## Argument Reference

- `pinto_provider` - (String, Optional) Provider that pinto will use to store DNS entries (Required if provider is not set globally for the terraform provider)
- `pinto_environment` - (String, Optional) Environment at the provider that will be used to sore DNS entries
- `zone` - (String, Required) The name of the zone
- `name` - (String, Required) Filter records per name (allowed values are "A", "NS", "CNAME", "SOA", "PTR", "MX", "TXT", "SRV", "AAAA", "SPF")
- `type` - (String, Required) Filter records per record type (allowed values are "IN", "CH", "HS", "CS")

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the record
- `ttl` - (Int) Time to Life of the record
- `class` - (String) Class of this record (allowed values are "IN", "CH", "HS", "CS")
- `data` - (String) Target-Server behind this record
