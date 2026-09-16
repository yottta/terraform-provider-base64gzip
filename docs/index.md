---
page_title: "Provider: base64gzip"
description: |-
  Provider exposing functions to provide legacy output since go1.27 changed the gzip algorithm
---

# Base64gzip

The "base64gzip" provider exposes 2 functions to allow users to use the old
behavior provided by OpenTofu pre 1.13 when using `base64gzip` and `base64gunzip`
functions.

Go's [`compress/flate`](https://pkg.go.dev/compress/flate) changed its default-compression-level
output between Go 1.26 and Go 1.27: for the same input, a Go 1.27 build emits a
different — equally valid, identically decompressing — Base64 string than a Go 1.26 build.
OpenTofu 1.12.x is built with go1.26.x, which contains the old gzip compression algorithm while
OpenTofu 1.13.x onwards is built with go1.27.x which changed the gzip algorithm which creates the
issue where a field like `aws_instance.user_data` will be flagged as updated
and OpenTofu might plan a replacement of the instance.

This provider is built using Go 1.26.x which uses the old gzip algorithm ensuring that the output
of the exposed functions match the output of the corresponded OpenTofu core functions pre 1.13.

The maintenance of this provider will end together with the support of Go 1.26.

## Usage
```terraform
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
```