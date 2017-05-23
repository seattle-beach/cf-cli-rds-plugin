package cf_rds_test

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/plugin/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/seattle-beach/cf-cli-rds-plugin/cf_rds"
)

var _ = Describe("CfRds", func() {

	Describe("BasicPlugin", func() {
		Context("with uri option", func() {
			var ui MockUi
			var conn *pluginfakes.FakeCliConnection
			var p *BasicPlugin
			var args []string

			BeforeEach(func() {
				conn = &pluginfakes.FakeCliConnection{}
				ui = MockUi{}

				p = &BasicPlugin{
					UI: &ui,
				}
				args = []string{"aws-rds", "create", "name", "--uri", "postgres://user:pwd@example.com:5432/database"}
			})

			JustBeforeEach(func() {
				p.Run(conn, args)
			})

			It("creates a user-provided service with user-provided RDS instance", func() {
				Expect(conn.CliCommandArgsForCall(0)).To(Equal([]string{"cups", "name", "-p", "{\"uri\":\"postgres://user:pwd@example.com:5432/database\"}"}))
			})

			Context("success message", func() {
				BeforeEach(func() {
					conn.GetCurrentSpaceReturns(plugin_models.Space{
						plugin_models.SpaceFields{
							Guid: "fake-guid",
							Name: "fake-space",
						},
					}, nil)
				})

				It("displays success message", func() {
					Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.Name}} in space {{.Space}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
					Expect(ui.Data).To(Equal(map[string]interface{}{
						"Name":  "name",
						"Space": "fake-space",
					}))
				})
			})
		})
	})
})

// test case for error from cli command
type MockUi struct {
	TextTemplate string
	Data map[string]interface{}
}

func (u *MockUi) DisplayText(template string, data ...map[string]interface{}) {
	u.TextTemplate = template
	u.Data = data[0]
}

func (u *MockUi) DisplayError(err error) {
	panic("NOT IMPLEMENTED")
}