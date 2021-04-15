---
page_title: "Data Source pinto_dns_records"
subcategory: "dns"
description: |-
  The records data source allows for the retrieval of existing records of a zone.
---

# Data Source `pinto_dns_records`

The records data source allows for the retrieval of existing records of a zone.

## Example Usage

```terraform
data "pinto_dns_records" "my_records" {
  pinto_provider = "digitalocean"
  zone           = "my.zone.com."
}
```

## Argument Reference

- `pinto_provider` - (String, Optional) Provider that pinto will use to store DNS entries (Required if provider is not set globally for the terraform provider)
- `pinto_environment` - (String, Optional) Environment at the provider that will be used to sore DNS entries
- `zone` - (String, Required) The name of the zone
- `name` - (String, Optional) Filter records per name
- `record_type` - (String, Optional) Filter records per record type

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the record collection
- `records` - (Records) A list of records in a zone. See [Records](#records) for details.

### Records

- `id` - (String) The id of the DNS record
- `name` - (String) The name of the DNS record
- `ttl` - (Int) Time to Life of the record
- `type` - (String) Type of this record (allowed values are "A", "NS", "CNAME", "SOA", "PTR", "MX", "TXT", "SRV", "AAAA", "SPF")
- `class` - (String) Class of this record (allowed values are "IN", "CH", "HS", "CS")
- `data` - (String) Target-Server behind this record
