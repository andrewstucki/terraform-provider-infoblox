# [Terraform](https://github.com/hashicorp/terraform) Infoblox Provider

This provider is a fork from https://github.com/prudhvitella/terraform-provider-infoblox since
use cases are diverging and beginning to make incompatible API changes from the original.

The Infoblox provider is used to interact with the
resources supported by Infoblox. The provider needs to be configured
with the proper credentials before it can be used.

##  Download
Download builds for Darwin, Linux and Windows from the [releases page](https://github.com/andrewstucki/terraform-provider-infoblox/releases/).

## Example Usage

```
# Configure the Infoblox provider
provider "infoblox" {
  username = "${var.infoblox_username}"
  password = "${var.infoblox_password}"
  host  = "${var.infoblox_host}"
  sslverify = "${var.infoblox_sslverify}"
  usecookies = "${var.infoblox_usecookies}"
}

# Create a record
resource "infoblox_record" "www" {
  ...
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The Infoblox username. It must be provided, but it can also be sourced from the `INFOBLOX_USERNAME` environment variable.
* `password` - (Required) The password associated with the username. It must be provided, but it can also be sourced from the `INFOBLOX_PASSWORD` environment variable.
* `host` - (Required) The base url for the Infoblox REST API, but it can also be sourced from the `INFOBLOX_HOST` environment variable.
* `sslverify` - (Required) Enable ssl for the REST api, but it can also be sourced from the `INFOBLOX_SSLVERIFY` environment variable.
* `usecookies` - (Optional) Use cookies to connect to the REST API, but it can also be sourced from the `INFOBLOX_USECOOKIES` environment variable

# infoblox\_record

Provides a Infoblox record resource.

## Example Usage

```
# Add a record to the domain
resource "infoblox_record" "foobar" {
  value = "192.168.0.10"
  name = "terraform"
  domain = "mydomain.com"
  type = "A"
  ttl = 3600
}
```

## Argument Reference

See [related part of Infoblox Docs](https://godoc.org/github.com/fanatic/go-infoblox) for details about valid values.

The following arguments are supported:

* `domain` - (Required) The domain to add the record to
* `value` - (Required) The value of the record; its usage will depend on the `type` (see below)
* `name` - (Required) The name of the record
* `ttl` - (Integer, Optional) The TTL of the record
* `type` - (Required) The type of the record

## DNS Record Types

The type of record being created affects the interpretation of the `value` argument.

#### A Record

* `value` is the hostname

#### CNAME Record

* `value` is the alias name

#### AAAA Record

* `value` is the IPv6 address

## Attributes Reference

The following attributes are exported:

* `domain` - The domain of the record
* `value` - The value of the record
* `name` - The name of the record
* `type` - The type of the record
* `ttl` - The TTL of the record

# infoblox\_ip

Queries the next available IP address from a network and returns it in a computed variable
that can be used by the infoblox_record resource.

## Example Usage

```
# Acquire the next available IP from a network CIDR
#it will create a variable called "ipaddress"
resource "infoblox_ip" "ip_addresses" {
  cidr = "10.0.0.0/24"
  ip_count = 2
}


# Add a record to the domain
resource "infoblox_record" "foo" {
  value = "${infoblox_ip.ip_addresses.addresses[0]}"
  name = "foo"
  domain = "mydomain.com"
  type = "A"
  ttl = 3600
}

# Add a record to the domain
resource "infoblox_record" "bar" {
  value = "${infoblox_ip.ip_addresses.addresses[1]}"
  name = "bar"
  domain = "mydomain.com"
  type = "A"
  ttl = 3600
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Required) The network to search for - example 10.0.0.0/24
* `ip_count` - (Optional, default 1) The number of ip addresses to request
* `exclude` - (Optional, default []) A list of ip addresses to exclude

The following values are returned:

* `addresses` - A list of ip addresses returned from infoblox
