# cf-cli-rds-plugin

Cloud Foundry Plugin to provision a Relational Database Service (RDS) instance
and connect it to a Pivotal Web Services (PWS) App.

## Getting Started

`go get github.com/seattle-beach/cf-cli-rds-plugin`

`cf install-plugin $GOPATH/bin/cf-cli-rds-plugin`

`cf aws`

For usage of the plugin, you can run:

`cf aws -h` or `cf aws --help`

## Running Tests
Install ginkgo on your machine. For instructions go to: `https://github.com/onsi/ginkgo`

Run `ginkgo` from the plugin path (`$GOPATH/src/github.com/seattle-beach/cf-cli-rds-plugin/`)