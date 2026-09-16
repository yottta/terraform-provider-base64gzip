terraform {
  required_providers {
    base64gzip = {
      source = "yottta/base64gzip"
    }
  }
}

locals {
  raw     = "hello world"
  encoded = "H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQAAAP//AQAA//+FEUoNCwAAAA=="
}

output "compressed" {
  value = "${local.raw} encodes to ${provider::base64gzip::base64gzip(local.raw)}"
}

output "decompressed" {
  value = "${local.encoded} decodes to ${provider::base64gzip::base64gunzip(local.encoded)}"
}