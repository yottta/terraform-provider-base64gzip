output "compressed" {
  value = provider::base64gzip::base64gzip("hello world")
}