package cf_rds_test

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/plugin/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/seattle-beach/cf-cli-rds-plugin/cf_rds"
	"github.com/seattle-beach/cf-cli-rds-plugin/cf_rds/fakes"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/seattle-beach/cf-cli-rds-plugin/api"
	"time"
	"code.cloudfoundry.org/cli/plugin"
)

var _ = Describe("CfRds", func() {
	Describe("BasicPlugin", func() {
		Describe("Run", func() {
			Context("register", func() {
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
					args = []string{"aws-rds-register", "name", "--uri", "postgres://user:pwd@example.com:5432/database"}
				})

				It("creates a user-provided service with user-provided RDS instance", func() {
					p.Run(conn, args)
					Expect(conn.CliCommandCallCount()).To(Equal(1))
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
						p.Run(conn, args)
						Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.ServiceName}} in space {{.Space}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
						Expect(ui.Data).To(Equal(map[string]interface{}{
							"ServiceName": "name",
							"Space":       "fake-space",
						}))
					})
				})

				Context("error cases", func() {
					It("returns an error if there are not enough arguments", func() {
						args = []string{"aws-rds-register", "name"}
						p.Run(conn, args)

						Expect(ui.Err).To(MatchError(ContainSubstring("cf aws-rds-register SERVICE_NAME --uri URI")))
					})

					It("returns an error if the --uri option flag is not provided", func() {
						args = []string{"aws-rds-register", "name", "--foo", "postgres://foo"}
						p.Run(conn, args)

						Expect(ui.Err).To(MatchError(ContainSubstring("cf aws-rds-register SERVICE_NAME --uri URI")))
					})
				})
			})

			Context("create", func() {
				var ui MockUi
				var conn *pluginfakes.FakeCliConnection
				var fakeApi *fakes.FakeApi
				var p *BasicPlugin
				var args []string
				var subnetGroup *rds.DBSubnetGroup

				BeforeEach(func() {
					conn = &pluginfakes.FakeCliConnection{}
					ui = MockUi{}
					fakeApi = &fakes.FakeApi{}

					p = &BasicPlugin{
						UI:           &ui,
						Api:          fakeApi,
						WaitDuration: time.Millisecond,
					}
					args = []string{"aws-rds-create", "name"}
					subnetGroup = &rds.DBSubnetGroup{
						DBSubnetGroupArn:         aws.String("arn:aws:rds:us-east-1:787194449165:subgrp:default-vpc-f7f7098e"),
						DBSubnetGroupDescription: aws.String("Created from the RDS Management Console"),
						DBSubnetGroupName:        aws.String("default-vpc-vpcid"),
						SubnetGroupStatus:        aws.String("Complete"),
						Subnets: []*rds.Subnet{
							{
								SubnetAvailabilityZone: &rds.AvailabilityZone{
									Name: aws.String("us-east-1d"),
								},
								SubnetIdentifier: aws.String("subnet-dc92a7b9"),
								SubnetStatus:     aws.String("Active"),
							}},
						VpcId: aws.String("vpcid"),
					}

					fakeApi.GetSubnetGroupsReturns([]*rds.DBSubnetGroup{subnetGroup}, nil)

					fakeApi.CreateInstanceStub = func(instance *api.DBInstance) chan error {
						errChan := make(chan error, 1)
						errChan <- nil

						instance.ResourceID = "resourceid"
						instance.SecGroups = []*rds.VpcSecurityGroupMembership{{
							VpcSecurityGroupId: aws.String("vpcgroup"),
						}}
						instance.ARN = "arn:aws:rds:us-east-1:10101010:db:name"
						instance.DBName = "database"
						instance.Password = "password"
						instance.DBURI = "postgres://root:password@test-uri.us-east-1.rds.amazonaws.com:5432/database"

						return errChan
					}
				})

				It("lists the DB subnet groups in the user's account", func() {
					p.Run(conn, args)
					Expect(fakeApi.GetSubnetGroupsCallCount()).To(Equal(1))
					Expect(fakeApi.CreateInstanceCallCount()).To(Equal(1))
					instance := fakeApi.CreateInstanceArgsForCall(0)

					Expect(instance.InstanceName).To(Equal("name"))
					Expect(instance.SubnetGroup).To(Equal(subnetGroup))
					Expect(instance.InstanceClass).To(Equal("db.t2.micro"))
					Expect(instance.Engine).To(Equal("postgres"))
					Expect(instance.Storage).To(Equal(int64(20)))
					Expect(instance.AZ).To(Equal("us-east-1a"))
					Expect(instance.Port).To(Equal(int64(5432)))
					Expect(instance.Username).To(Equal("root"))
				})

				It("creates a user-provided service with the created RDS instance", func() {
					p.Run(conn, args)
					Expect(conn.CliCommandCallCount()).To(Equal(2))

					cupsArgs := conn.CliCommandArgsForCall(0)
					Expect(len(cupsArgs)).To(Equal(4))
					Expect(cupsArgs[0:3]).To(Equal([]string{"cups", "name", "-p"}))
					Expect(cupsArgs[3]).To(MatchJSON(`{
					"instance_id": "name",
					"username": "root"
				}`))
				})

				It("captures the uri and calls uups", func() {
					p.Run(conn, args)
					Expect(conn.CliCommandCallCount()).To(Equal(2))

					uupsArgs := conn.CliCommandArgsForCall(1)
					Expect(len(uupsArgs)).To(Equal(4))
					Expect(uupsArgs[0:3]).To(Equal([]string{"uups", "name", "-p"}))
					Expect(uupsArgs[3]).To(MatchJSON(`{
					"instance_id": "name",
					"arn": "arn:aws:rds:us-east-1:10101010:db:name",
					"resource_id": "resourceid",
					"uri": "postgres://root:password@test-uri.us-east-1.rds.amazonaws.com:5432/database",
					"username": "root",
					"password": "password",
					"database": "database"
				}`))
				})

				It("prints a success message", func() {
					p.Run(conn, args)
					Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.ServiceName}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
					Expect(ui.Data["ServiceName"]).To(Equal("name"))
					Expect(ui.Data["RDSID"]).To(Equal("resourceid"))
					Expect(ui.Data["VPC"]).To(Equal("vpcid"))
					Expect(ui.Data["SecGroup"]).To(Equal("vpcgroup"))
				})

			})

			Context("refresh", func() {
				var ui MockUi
				var conn *pluginfakes.FakeCliConnection
				var fakeApi *fakes.FakeApi
				var p *BasicPlugin
				var args []string

				BeforeEach(func() {
					conn = &pluginfakes.FakeCliConnection{}
					ui = MockUi{}
					fakeApi = &fakes.FakeApi{}

					p = &BasicPlugin{
						UI:           &ui,
						Api:          fakeApi,
						WaitDuration: time.Millisecond,
					}
					args = []string{"aws-rds-refresh", "name"}

					errChan := make(chan error, 1)
					errChan <- nil
					fakeApi.RefreshInstanceStub = func(instance *api.DBInstance) chan error {
						errChan := make(chan error, 1)
						errChan <- nil

						instance.ARN = "arn:aws:rds:us-east-1:10101010:db:name"
						instance.ResourceID = "resourceid"
						instance.Username = "root"
						instance.DBName = "database"
						instance.SecGroups = []*rds.VpcSecurityGroupMembership{{
							VpcSecurityGroupId: aws.String("vpcgroup"),
						}}
						instance.SubnetGroup = &rds.DBSubnetGroup{
							VpcId: aws.String("vpcid"),
						}
						instance.Engine = "postgres"
						instance.DBURI = "postgres://root:password2@test-uri.us-east-1.rds.amazonaws.com:5432/database"
						instance.Password = "password2"

						return errChan
					}
				})

				It("recaptures DB info from AWS and resets the password", func() {
					p.Run(conn, args)
					Expect(fakeApi.RefreshInstanceCallCount()).To(Equal(1))
					instance := fakeApi.RefreshInstanceArgsForCall(0)

					Expect(instance.InstanceName).To(Equal("name"))
				})

				It("captures the uri and calls uups", func() {
					p.Run(conn, args)
					Expect(conn.CliCommandCallCount()).To(Equal(1))

					uupsArgs := conn.CliCommandArgsForCall(0)
					Expect(uupsArgs[0:3]).To(Equal([]string{"uups", "name", "-p"}))
					Expect(uupsArgs[3]).To(MatchJSON(`{
					"instance_id": "name",
					"arn": "arn:aws:rds:us-east-1:10101010:db:name",
					"resource_id": "resourceid",
					"uri": "postgres://root:password2@test-uri.us-east-1.rds.amazonaws.com:5432/database",
					"username": "root",
					"password": "password2",
					"database": "database"
				}`))
				})

				It("prints a success message", func() {
					p.Run(conn, args)
					Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.ServiceName}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
					Expect(ui.Data["ServiceName"]).To(Equal("name"))
					Expect(ui.Data["RDSID"]).To(Equal("resourceid"))
					Expect(ui.Data["VPC"]).To(Equal("vpcid"))
					Expect(ui.Data["SecGroup"]).To(Equal("vpcgroup"))
				})
			})
		})

		Describe("GetMetadata", func() {
			var ui MockUi
			var p *BasicPlugin

			BeforeEach(func() {
				ui = MockUi{}

				p = &BasicPlugin{
					UI: &ui,
				}
			})

			It("returns metadata for the plugin", func() {
				Expect(p.GetMetadata()).To(Equal(plugin.PluginMetadata{
					Name: "aws-plugin",
					Version: plugin.VersionType{
						Major: 1,
						Minor: 0,
						Build: 0,
					},
					MinCliVersion: plugin.VersionType{
						Major: 6,
						Minor: 7,
						Build: 0,
					},
					Commands: []plugin.Command{
						{
							Name:     "aws-rds-register",
							HelpText: "command to register existing RDS instance as a service with CF",

							UsageDetails: plugin.Usage{
								Usage: "cf aws-rds-register SERVICE_NAME --uri URI",
							},
						},
						{
							Name:     "aws-rds-create",
							HelpText: "command to create an RDS instance and register it as a service with CF",

							UsageDetails: plugin.Usage{
								Usage: "cf aws-rds-create SERVICE_NAME",
							},
						},
						{
							Name:     "aws-rds-refresh",
							HelpText: "command to update an existing RDS instance and register it as a service with CF (used in case the user quits rds-create command before the instance is fully available)",

							UsageDetails: plugin.Usage{
								Usage: "cf aws-rds-refresh SERVICE_NAME",
							},
						},
					},
				}))

			})
		})
	})
})

// test case for error from cli command
type MockUi struct {
	TextTemplate string
	Err          error
	Data         map[string]interface{}
}

func (u *MockUi) DisplayText(template string, data ...map[string]interface{}) {
	u.TextTemplate = template
	if data != nil {
		u.Data = data[0]
	}
}

func (u *MockUi) DisplayError(err error) {
	u.Err = err
}
