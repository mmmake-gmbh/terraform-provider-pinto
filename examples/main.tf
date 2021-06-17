terraform {
  required_providers {
    pinto = {
      source = "terraform.local/camaoag/project-pinto"
    }
  }
}

provider "pinto" {
  base_url       = "https://pinto.irgendwo.co"
  token_url      = "https://auth.pinto.irgendwo.co/connect/token"
  client_id      = "client_id"
  client_secret  = "super_secret"
  client_scope   = "openapigateway,nexus"
  pinto_provider = "digitalocean"
  pinto_environment = "prod1"
}

data "pinto_dns_zone" "zone1" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  name              = "env0.co."
}

data "pinto_dns_zones" "zones" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
}

data "pinto_dns_records" "records_env0" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  zone              = "env0.co."
}

data "pinto_dns_record" "record" {
  pinto_provider    = "digitalocean"
  pinto_environment = "prod1"
  zone              = "env0.co."
  name              = "pinto"
  type              = "A"
}

//resource "pinto_dns_record" "test_record1" {
//  pinto_provider    = "digitalocean"
//  pinto_environment = "prod1"
//  zone              = "env0.co."
//  name              = "testrecord"
//  type              = "TXT"
//  class             = "IN"
//  data              = "127.0.0.1"
//  ttl               = 1800
//}

//resource "pinto_dns_zone" "test_zone1" {
//  pinto_provider    = "digitalocean"
//  pinto_environment = "prod1"
//  name              = "test.purrrfect."
//}

//resource "pinto_dns_zone" "testi" {
//  pinto_environment = "prod1"
//}

//resource "pinto_dns_record" "testi2" {
//  pinto_environment = "prod1"
//}
