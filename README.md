# cf-cli-rds-plugin

Cloud Foundry Plugin to provision a Relational Database Service (RDS) instance
and connect it to a Pivotal Web Services (PWS) App.

## Getting Started

1. Make sure `$GOPATH` is set. If it isn't already set, `$HOME/go` is probably a reasonable value to use.
2. `go get github.com/seattle-beach/cf-cli-rds-plugin`
3. `cf install-plugin $GOPATH/bin/cf-cli-rds-plugin`
4. `cf aws`

For usage of the plugin, you can run:

`cf aws -h` or `cf aws --help`

## Running Tests
Install ginkgo on your machine. For instructions go to: `https://github.com/onsi/ginkgo`

If the `ginkgo` command did not get installed, try `go get github.com/onsi/ginkgo/ginkgo`

Run `ginkgo` from the plugin path (`$GOPATH/src/github.com/seattle-beach/cf-cli-rds-plugin/`)