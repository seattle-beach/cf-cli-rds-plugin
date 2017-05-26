# cf-cli-rds-plugin

Cloud Foundry Plugin to provision a Relational Database Service (RDS) instance
and connect it to a Pivotal Web Services (PWS) App.

It exposes three commands:
1. `cf aws-rds-create SERVICE_NAME` - create an RDS instance and register it as a service with CF
1. `cf aws-rds-register SERVICE_NAME --uri URI` - register existing RDS instance as a service with CF
1. `cf aws-rds-refresh SERVICE_NAME` - update an existing RDS instance and register it as a service with CF (used in case the user quits aws-rds-create command before the instance is fully available)

## Getting Started

### Building from source
1. Make sure `$GOPATH` is set. If it isn't already set, `$HOME/go` is probably a reasonable value to use. Make sure `$GOPATH/bin` is added to your system PATH.
1. `go get github.com/seattle-beach/cf-cli-rds-plugin`
1. `mv $GOPATH/bin/cf-cli-rds-plugin $GOPATH/bin/aws-plugin`
1. `cf install-plugin $GOPATH/bin/aws-plugin`
1. `cf aws-rds-create`/`cf aws-rds-register`

### Downloading plugin binary directly
1. Navigate to the github releases page of this repository.
1. Download the `aws-plugin` binary.
1. `cf install-plugin <path/to/aws-plugin>`
1. `cf aws-rds-create`/`cf aws-rds-register`

For usage of the plugin, you can run:

`cf aws-rds-create -h` or `cf aws-rds-create --help` (or any of the commands followed by --help)

## Running Tests
Install ginkgo on your machine. For instructions go to: `https://github.com/onsi/ginkgo`

If the `ginkgo` command did not get installed, try `go get github.com/onsi/ginkgo/ginkgo`

Run `ginkgo -r .` from the plugin path (`$GOPATH/src/github.com/seattle-beach/cf-cli-rds-plugin/`)

## Uninstalling
`cf uninstall-plugin aws-plugin`
