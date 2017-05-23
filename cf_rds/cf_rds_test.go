package cf_rds_test

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/seattle-beach/cf-cli-rds-plugin/cf_rds"
)

var _ = Describe("CfRds", func() {

	Describe("BasicPlugin", func() {
		FContext("with uri option", func() {
			var ui MockUi
			var conn *pluginfakes.FakeCliConnection

			BeforeEach(func() {
				conn = &pluginfakes.FakeCliConnection{}
				ui = MockUi{}

				p := &BasicPlugin{
					UI: &ui,
				}
				args := []string{"aws-rds", "create", "name", "--uri", "postgres://user:pwd@example.com:5432/database"}
				p.Run(conn, args)
			})

			It("creates a user-provided service with user-provided RDS instance", func() {
				Expect(conn.CliCommandArgsForCall(0)).To(Equal([]string{"cups", "name", "-p", "{\"uri\":\"postgres://user:pwd@example.com:5432/database\"}"}))
			})

			It("displays success message", func() {
				Expect(ui.TextTemplate).To(Equal("SUCCESS"))
			})
		})
	})
})

// test case for error from cli command
type MockUi struct {
	TextTemplate string
}

func (u *MockUi) DisplayText(template string, data ...map[string]interface{}) {
	u.TextTemplate = template
}

func (u *MockUi) DisplayError(err error) {
	panic("NOT IMPLEMENTED")
}