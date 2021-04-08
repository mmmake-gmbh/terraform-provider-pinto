---
page_title: "Resource pinto_dns_record"
subcategory: "dns"
description: |-
  The pinto_dns_record resource allows for the creation of DNS records.
---

# Resource `pinto_dns_record`

The pinto_dns_record resource allows for the creation of DNS records.

## Example Usage

```terraform
resource "pinto_dns_record" "example_record" {
  zone = "env0.co."
  name = "testrecord"
  type = "A"
  class = "IN"
  data = "127.0.0.1"
}
```

## Argument Reference

- `zone` - (String, Required) The zone in which the DNS record is to be created
- `name` - (String, Required) The name of the DNS record
- `type` - (String, Required) The type of the DNS record (allowed values are "A", "NS", "CNAME", "SOA", "PTR", "MX", "TXT", "SRV", "AAAA", "SPF")
- `class` - (String, Required) The class of the DNS record (allowed values are "IN", "CH", "HS", "CS")
- `data` - (String, Required) The target locations the DNS record is pointing to
- `ttl` - (Integer, Optional) TTL of the record (Default: 3600)

## Attributes Reference

In addition to all the arguments above, the following attributes are exported.

- `id` - (String) The ID of the DNS record
