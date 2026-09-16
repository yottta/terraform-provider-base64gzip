# `base64gzip` and `base64gunzip` as provider functions

This is a small, stateless provider that exposes only 2 functions:

* `base64gzip` — compresses a string with gzip and encodes the result in Base64.
* `base64gunzip` — the inverse: Base64-decodes a string and decompresses it with gzip.

## Purpose

The goal is to emit *byte-identical* output to OpenTofu's pre-1.13 built-in `base64gzip`/`base64gunzip`, so that this
provider can be used as an intermediary step when migrating to OpenTofu 1.13. 
You can move a configuration onto `provider::base64gzip::base64gzip(...)` and later swap it for the built-in `base64gzip(...)`
when the resources affected by this change could and will be recreated with OpenTofu 1.13.

Go's `compress/flate` changed its default-compression-level output between Go 1.26 and Go 1.27: for the same input, a Go 1.27 build emits a
different — equally valid, identically decompressing — Base64 string than a Go 1.26 build. OpenTofu 1.12.x is
built with go1.26.x, which contains the old gzip compression algorithm while OpenTofu 1.13.x onwards is built with go1.27.x which
changed the gzip algorithm which creates the issue where a field like `aws_instance.user_data` will be flagged as updated
and OpenTofu might plan a replacement of the instance.

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
OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
WORKDIR="$(mktemp -d)"
DIR="${WORKDIR}/plugins/registry.opentofu.org/yottta/base64gzip/${VERSION}/${OS_ARCH}"

prev_dir="$(pwd)"
mkdir -p "${DIR}"
echo "work dir ${WORKDIR}"
GOTOOLCHAIN=go1.26.8 go build -o "${DIR}/terraform-provider-base64gzip_v${VERSION}" .

cat > "${WORKDIR}/mirror.tfrc" <<EOF
provider_installation {
  filesystem_mirror { path = "${WORKDIR}/plugins" }
}
EOF
cat > "${WORKDIR}/main.tf" <<EOF
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
  value = "\${local.raw} encodes to \${provider::base64gzip::base64gzip(local.raw)}"
}

output "decompressed" {
  value = "\${local.encoded} decodes to \${provider::base64gzip::base64gunzip(local.encoded)}"
}
EOF


export TF_CLI_CONFIG_FILE="${WORKDIR}/mirror.tfrc"
cd "${WORKDIR}"
tofu init
tofu apply -auto-approve

cd "${prev_dir}"
```
