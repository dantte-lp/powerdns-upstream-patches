# terraform-provider-powerdns

> **This is a fork** of [`mmianl/terraform-provider-powerdns`](https://github.com/mmianl/terraform-provider-powerdns),
> taken at `0dac0e7` (v2.3.0), being moved to the Terraform Plugin Framework and
> a full standards set. Contributions intended for upstream are opened against
> that repository, not this one — see [`docs/release.md`](docs/release.md).
>
> **Contributors and agents start at [`AGENTS.md`](AGENTS.md).**
> Documentation index: [`docs/README.md`](docs/README.md).
> Coverage analysis and the defect register live in the sibling
> `powerdns-capability-map` repository.

The Terraform PowerDNS provider allows you to manage PowerDNS zones, records, views, and networks using Terraform. It is maintained by mmianl.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.11.x
- [Go](https://golang.org/doc/install) >=1.24.x (to build the provider plugin)
- [Goreleaser](https://goreleaser.com) >=v6.3.x (for releasing provider plugin)

The Go ang Goreleaser minimum versions were set to be able to build plugin for Darwin/ARM64 architecture [see goreleaser notes.](https://goreleaser.com/deprecations/#builds-for-darwinarm64)

## Using the Provider

```hcl
terraform {
  required_providers {
    powerdns = {
      source = "mmianl/powerdns"
      version = "1.8.1"
    }
  }
}

provider "powerdns" {
  server_url = "https://host:port/"           # authoritative server url (can also be provided with PDNS_SERVER_URL variable)
  recursor_server_url = "https://host:port/"  # recursor server url (can also be provided with PDNS_RECURSOR_SERVER_URL variable)
  api_key             = "secret"              # can also be provided with PDNS_API_KEY variable
}

# Note: The provider supports both PowerDNS Authoritative Server and PowerDNS Recursor.
# Configure server_url for authoritative operations and recursor_server_url for recursor operations.
```

### Supported authoritative resources

- `powerdns_zone`
- `powerdns_zone_metadata`
- `powerdns_record`
- `powerdns_record_soa`
- `powerdns_ptr_record`
- `powerdns_reverse_zone`
- `powerdns_view`
- `powerdns_network`

For detailed usage see [provider's documentation page](https://registry.terraform.io/providers/mmianl/powerdns/latest/docs)

## Environment Variables

The provider supports configuration via environment variables as an alternative to the provider block configuration:

- `PDNS_SERVER_URL` - The URL of the PowerDNS Authoritative Server (e.g., `https://host:port/`)
- `PDNS_API_KEY` - The API key for authenticating with the PowerDNS server
- `PDNS_RECURSOR_SERVER_URL` - The URL of the PowerDNS Recursor Server (e.g., `https://host:port/`)

When these environment variables are set, you can use the provider without explicit configuration:

```hcl
provider "powerdns" {}
```

## Building The Provider

Clone the provider repository:

```sh
$ git clone git@github.com:mmianl/terraform-provider-powerdns.git

Navigate to repository directory:

```sh
$ cd terraform-provider-powerdns
```

Build repository:

```sh
$ go build
```

This will compile and place the provider binary, `terraform-provider-powerdns`, in the current directory.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is _recommended_).
You'll also need to have `$GOPATH/bin` in your `$PATH`.

Make sure the changes you performed pass linting:

```sh
$ task lint
```

To install the provider, run `task build`. This will build the provider and put the provider binary in the current working directory.

```sh
$ task build
```

In order to run local provider tests, you can simply run `task test`.

```sh
$ task test
```

For running acceptance tests locally, you'll need to use `docker-compose` to prepare the test environment:

```sh
docker-compose run --rm setup
```

After setup is done, run the acceptance tests with `task testacc` (note the env variables needed to interact with the PowerDNS container)

- HTTP

```sh
~$  PDNS_SERVER_URL=http://localhost:8081 \
    PDNS_API_KEY=secret \
    task testacc
```

- HTTPS

```sh
~$  PDNS_SERVER_URL=localhost:4443 \
    PDNS_API_KEY=secret \
    PDNS_CACERT=$(cat ./tests/files/ssl/rootCA/rootCA.crt) \
    task testacc
```

And finally cleanup containers spun up by `docker-compose`:

```sh
~$ docker-compose down
```
