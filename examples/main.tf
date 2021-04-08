terraform {
  required_providers {
    pinto = {
      source = "terraform.local/whizus/pinto"
    }
  }
}

provider "pinto" {
  client_id = "machineclient"
  client_secret = "Secret123$"
  client_scope = "openapigateway,nexus"
  provider = "digitalocean"
  environment = "prod1"
}

data "pinto_dns_zone" "zone1" {
  name = "env0.co."
}

data "pinto_dns_zones" "zones" {}

data "pinto_dns_records" "records_env0" {
  zone = "env0.co."
}

//resource "pinto_dns_record" "test_record1" {
//  zone = "env0.co."
//  name = "testrecord"
//  type = "A"
//  class = "IN"
//  data = "127.0.0.1"
//  ttl  = 1800
//}
//
//resource "pinto_dns_zone" "test_zone1" {
//  name = "test.purrfect."
//}

resource "pinto_dns_zone" "testi" {
  name = "truck.duck."
}