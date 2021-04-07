---
page_title: "Data Source dns_records"
subcategory: "dns"
description: |-
  The records data source allows for retrieval of existing records of a zone.
---

# Data Source `dns_records`

The records data source allows for retrieval of existing records of a zone.

## Example Usage

```terraform
data "pinto_dns_records" "my_records" {
  zone = "my.zone.com."
}
```

## Argument Reference

- `zone` - (String, Required) The name of the zone
- `name` - (String, Optional) Filter records per name
- `record_type` - (String, Optional) Filter records per record type

## Attributes Reference

- `id` - (String) The ID of the record collection
- `records` - (Records) A list of records in a zone. See [Records](#records) for details.

### Records

- `name` - (String) The name of the dns record
- `ttl` - (Int) Time to Life of the record
- `record_type` - (String) Type of this record (A, TXT, AAAA, ...)
- `class` - (String) Class of this record
- `data` - (String) Target-Server behind this record
