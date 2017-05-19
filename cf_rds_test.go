package main_test

import (
	 . "github.com/seattle-beach/cf-cli-rds-plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"

)

var _ = Describe("CfRds", func() {
	Describe("BasicPlugin", func() {
		Context("with uri option", func() {
			It("creates a user-provided service with user-provided RDS instance", func() {
				conn := &pluginfakes.FakeCliConnection{}
				p := &BasicPlugin{}
				args := []string { "aws-rds", "create", "name", "--uri", "postgres://user:pwd@example.com:5432/database" }
				p.Run(conn, args)
				cliCommandArgs := conn.CliCommandArgsForCall(0)
				Expect(cliCommandArgs).To(Equal([]string {"cups", "name", "-p", "{\"uri\":\"postgres://user:pwd@example.com:5432/database\"}"}))
			})
		})
	})
})
