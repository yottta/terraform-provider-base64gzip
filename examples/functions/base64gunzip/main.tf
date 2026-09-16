output "decompressed" {
  value = provider::base64gzip::base64gunzip("H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQAAAP//AQAA//+FEUoNCwAAAA==")
}