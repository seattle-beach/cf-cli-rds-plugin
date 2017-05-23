package cf_rds_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfCliRdsPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfCliRdsPlugin Suite")
}

