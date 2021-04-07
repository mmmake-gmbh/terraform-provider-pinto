---
page_title: "Pinto Provider"
subcategory: ""
description: |-
  Terraform provider for interacting with Pinto API.
---

# Pinto Provider

The Pinto Provider can be used to interact with [Pinto](https://pinto.irgendwo.co/api/dns/swagger/index.html).

## Example Usage

Do not keep your authentication password in HCL for production environments, use Terraform environment variables.

```terraform
provider "pinto" {
  provider      = "digitalocean"
  environment   = "prod1"
  client_id     = "client"
  client_secret = "test123"
  client_scope  = "openapigateway,nexus"
}
```

## Argument Reference

The following arguments are supported:

- `provider` - (String, Required) Provider that pinto will use to store DNS entries
- `environment` - (String, Required) Environment at the provider that will be used to sore DNS entries
- `client_id` - (String, Optional) Client ID for client-credentials authentication (either this or API Key has to be set)
- `client_secret` - (String, Optional) Client Secret for client-credentials authentication
- `client_scope` - (String, Optional) Has to be "openapigateway,nexus"
- `api_key` - (String, Optional) API-Key to interact with Pinto (either this or client-credentials has to be set)
- `base_url` - (String, Optional) Overwrite the base-url of Pinto (Default: [https://pinto.irgendwo.co](https://pinto.irgendwo.co))
- `token_url` - (String, Optional) Overrite token-url for authentication (Default: [https://auth.pinto.irgendwo.co/connect/token](https://auth.pinto.irgendwo.co/connect/token))
