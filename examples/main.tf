terraform {
  required_providers {
    pinto = {
      source = "terraform.local/camaoag/pinto"
    }
  }
}

provider "pinto" {
  base_url       = "https://test.camao.domains.fascicularis.de"
  token_url      = "https://auth.test.camao.domains.fascicularis.de/connect/token"
  client_id      = ""
  client_secret  = ""
  client_scope   = ""
  credentials_id = ""
  pinto_provider = "digitalocean"
  pinto_environment = "prod1"
}

# testing zone
data "pinto_dns_zone" "testing_zone" {
  name              = "env0.co."
}

data "pinto_dns_records" "testing_zone_records" {
  zone              = data.pinto_dns_zone.testing_zone.name
}

resource "pinto_dns_zone" "created_zone" {
  name              = "test.purrrrfect."
}

resource "pinto_dns_record" "test_record_2" {
  zone              = pinto_dns_zone.created_zone.name
  name              = "testrecord"
  type              = "TXT"
  class             = "IN"
  data              = "127.0.0.1"
  ttl               = 1800
}
