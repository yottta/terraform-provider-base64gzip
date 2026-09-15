# `base64gzip` and `base64gunzip` as provider functions

This is a small, stateless provider that exposes only 2 functions:

* `base64gzip` — compresses a string with gzip and encodes the result in Base64.
* `base64gunzip` — the inverse: Base64-decodes a string and decompresses it with gzip.

## Purpose

The goal is to emit *byte-identical* output to OpenTofu's pre-v1.13 built-in `base64gzip`/`base64gunzip`, so that this
provider can be used as an intermediary step when migrating to OpenTofu v1.13. 
You can move a configuration onto `provider::base64gzip::base64gzip(...)` and later swap it for the built-in `base64gzip(...)`
when the resources affected by this change when upgrading to OpenTofu v1.13.

That parity depends on the Go toolchain, not just on the algorithm. Go's `compress/flate` changed its
default-compression-level output between Go 1.26 and Go 1.27: for the same input, a Go 1.27 build emits a
different — equally valid, identically decompressing — Base64 string than a Go 1.26 build. OpenTofu 1.12.x is
built with go1.26.z, so **this provider must be built with a go1.26.x toolchain** for the guarantee to hold.
Built with go1.27, it still round-trips correctly, but it is no longer a drop-in match for the core functions
and swapping to the built-in would produce a diff.

## Provider configuration

```hcl
terraform {
  required_providers {
    base64gzip = {
      source = "yottta/base64gzip"
    }
  }
}
```

The provider takes no configuration arguments, so a `provider "base64gzip" {}` block is not required.

## Functions

### `base64gzip`

```hcl
output "compressed" {
  value = provider::base64gzip::base64gzip("hello world")
}
# => "H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQAAAP//AQAA//+FEUoNCwAAAA=="
```

#### Arguments

* `str` (`string`, required) — the string to compress and encode.

#### Return

A `string` containing the gzip-compressed input, Base64-encoded using the standard alphabet (with padding).

### `base64gunzip`

```hcl
output "decompressed" {
  value = provider::base64gzip::base64gunzip("H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQAAAP//AQAA//+FEUoNCwAAAA==")
}
# => "hello world"
```

#### Arguments

* `str` (`string`, required) — a Base64-encoded, gzip-compressed string.

#### Return

A `string` containing the decompressed data.

#### Errors

The function returns an argument error if the input is not valid standard-alphabet Base64
(`failed to decode base64 data: ...`) or if the decoded bytes are not a valid gzip stream
(`failed to read gzip data: ...`).

## Development

Build the provider (note the explicit toolchain — see [Purpose](#purpose)):

```shell
VERSION=0.0.1
OS_ARCH=darwin_arm64
DIR="$PWD/plugins/registry.opentofu.org/yottta/base64gzip/$VERSION/$OS_ARCH"

mkdir -p "$DIR"
GOTOOLCHAIN=go1.26.8 go build -o "$DIR/terraform-provider-base64gzip_v$VERSION" .

cat > mirror.tfrc <<EOF
provider_installation {
  filesystem_mirror { path = "$PWD/plugins" }
}
EOF

export TF_CLI_CONFIG_FILE="$PWD/mirror.tfrc"
tofu init
tofu apply -auto-approve
```
