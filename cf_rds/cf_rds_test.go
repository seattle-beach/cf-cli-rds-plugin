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
	"time"
)

var _ = Describe("CfRds", func() {

	Describe("BasicPlugin", func() {
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
				args = []string{"aws-rds", "register", "name", "--uri", "postgres://user:pwd@example.com:5432/database"}
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
					Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.Name}} in space {{.Space}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
					Expect(ui.Data).To(Equal(map[string]interface{}{
						"Name":  "name",
						"Space": "fake-space",
					}))
				})
			})

			Context("error cases", func() {
				It("returns an error if there are not enough arguments", func() {
					args = []string{"aws-rds", "register", "name"}
					p.Run(conn, args)

					Expect(ui.Err).To(MatchError(ContainSubstring("cf aws-rds register NAME --uri URI")))
				})

				It("returns an error if the --uri option flag is not provided", func() {
					args = []string{"aws-rds", "register", "name", "--foo", "postgres://foo"}
					p.Run(conn, args)

					Expect(ui.Err).To(MatchError(ContainSubstring("cf aws-rds register NAME --uri URI")))
				})
			})
		})

		Context("create", func() {
			var ui MockUi
			var conn *pluginfakes.FakeCliConnection
			var fakeRDSSvc *fakes.FakeRDSService
			var p *BasicPlugin
			var args []string

			BeforeEach(func() {
				conn = &pluginfakes.FakeCliConnection{}
				ui = MockUi{}
				fakeRDSSvc = &fakes.FakeRDSService{}

				p = &BasicPlugin{
					UI: &ui,
					Svc: fakeRDSSvc,
					WaitDuration: time.Millisecond,
				}
				args = []string{"aws-rds", "create", "name"}

				fakeRDSSvc.DescribeDBSubnetGroupsReturns(&rds.DescribeDBSubnetGroupsOutput{
					DBSubnetGroups: []*rds.DBSubnetGroup{{
						DBSubnetGroupArn: aws.String("arn:aws:rds:us-east-1:787194449165:subgrp:default-vpc-f7f7098e"),
						DBSubnetGroupDescription: aws.String("Created from the RDS Management Console"),
						DBSubnetGroupName: aws.String("default-vpc-vpcid"),
						SubnetGroupStatus: aws.String("Complete"),
						Subnets: []*rds.Subnet{
						{
							SubnetAvailabilityZone: &rds.AvailabilityZone {
								Name: aws.String("us-east-1d"),
							},
							SubnetIdentifier: aws.String("subnet-dc92a7b9"),
							SubnetStatus: aws.String("Active"),
						}},
						VpcId: aws.String("vpcid"),
					}},
				}, nil)

				fakeRDSSvc.CreateDBInstanceReturns(&rds.CreateDBInstanceOutput{
					DBInstance: &rds.DBInstance{
						DbiResourceId: aws.String("resourceid"),
						DBInstanceArn: aws.String("arn:aws:rds:us-east-1:10101010:db:name"),
						VpcSecurityGroups: []*rds.VpcSecurityGroupMembership{{
							VpcSecurityGroupId: aws.String("vpcgroup"),
						}},
					},

				}, nil)

				fakeRDSSvc.DescribeDBInstancesStub = func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
					if fakeRDSSvc.DescribeDBInstancesCallCount() < 5 {
						return &rds.DescribeDBInstancesOutput{
							DBInstances: []*rds.DBInstance{{
								DBInstanceIdentifier: aws.String("resourceid"),
								DBInstanceStatus: aws.String("creating"),
							}},
						}, nil
					}

					return &rds.DescribeDBInstancesOutput{
						DBInstances: []*rds.DBInstance{{
							DBInstanceIdentifier: aws.String("resourceid"),
							DBInstanceStatus: aws.String("available"),
							Endpoint: &rds.Endpoint {
								Port: aws.Int64(5432),
								Address: aws.String("test-uri.us-east-1.rds.amazonaws.com"),
							},
						}},
					}, nil
				}

				GenerateRandomAlphanumericString = func() string {
					return "password"
				}

				GenerateRandomString = func() string {
					return "database"
				}
			})

			It("lists the DB subnet groups in the user's account", func() {
				p.Run(conn, args)
				Expect(fakeRDSSvc.DescribeDBSubnetGroupsCallCount()).To(Equal(1))
				Expect(fakeRDSSvc.DescribeDBSubnetGroupsArgsForCall(0)).To(Equal(&rds.DescribeDBSubnetGroupsInput{
					Filters: []*rds.Filter{
						{
							Name: aws.String("*"),
							Values: []*string{
								aws.String("*"),
							},
						},
					},
					Marker:     aws.String("String"),
					MaxRecords: aws.Int64(20),
				}))
			})

			It("creates an RDS DB instance in the first subnet group", func() {
				p.Run(conn, args)
				Expect(fakeRDSSvc.CreateDBInstanceCallCount()).To(Equal(1))

				createDBInstanceInput := fakeRDSSvc.CreateDBInstanceArgsForCall(0)
				Expect(createDBInstanceInput.DBInstanceClass).To(Equal(aws.String("db.t2.micro")))
				Expect(createDBInstanceInput.DBInstanceIdentifier).To(Equal(aws.String("name")))
				Expect(createDBInstanceInput.Engine).To(Equal(aws.String("postgres")))
				Expect(createDBInstanceInput.AllocatedStorage).To(Equal(aws.Int64(20)))
				Expect(createDBInstanceInput.AvailabilityZone).To(Equal(aws.String("us-east-1a")))
				Expect(createDBInstanceInput.DBSubnetGroupName).To(Equal(aws.String("default-vpc-vpcid")))
				Expect(createDBInstanceInput.Port).To(Equal(aws.Int64(5432)))
			})

			It("creates a user-provided service with the created RDS instance", func() {
				p.Run(conn, args)
				Expect(conn.CliCommandCallCount()).To(Equal(2))

				cupsArgs := conn.CliCommandArgsForCall(0)
				Expect(len(cupsArgs)).To(Equal(4))
				Expect(cupsArgs[0:3]).To(Equal([]string{"cups", "name", "-p"}))
				Expect(cupsArgs[3]).To(MatchJSON(`{
					"instance_id": "name",
					"arn": "arn:aws:rds:us-east-1:10101010:db:name",
					"resource_id": "resourceid",
					"username": "root",
					"password": "password",
					"database": "database"
				}`))
			})

			It("polls for the db instance to be available", func() {
				p.Run(conn, args)
				Expect(fakeRDSSvc.DescribeDBInstancesCallCount()).To(Equal(6))
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
				Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
				Expect(ui.Data["Name"]).To(Equal("name"))
				Expect(ui.Data["RDSID"]).To(Equal("resourceid"))
				Expect(ui.Data["VPC"]).To(Equal("vpcid"))
				Expect(ui.Data["SecGroup"]).To(Equal("vpcgroup"))
			})

			Context("error cases", func() {
				Context("when there are no DB subnet groups", func() {
					BeforeEach(func() {
						fakeRDSSvc.DescribeDBSubnetGroupsReturns(&rds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*rds.DBSubnetGroup{},
						}, nil)
					})

					It("should return an error", func() {
						p.Run(conn, args)
						Expect(ui.Err).To(MatchError("Error: did not find any DB subnet groups to create RDS instance in"))
					})
				})

				Context("when there are no vpc security groups", func() {
					BeforeEach(func() {
						fakeRDSSvc.CreateDBInstanceReturns(&rds.CreateDBInstanceOutput{
							DBInstance: &rds.DBInstance{
								DbiResourceId: aws.String("resourceid"),
								DBInstanceArn: aws.String("resourcearn"),
								VpcSecurityGroups: []*rds.VpcSecurityGroupMembership{},
							},
						}, nil)
					})

					It("should return an error", func() {
						p.Run(conn, args)
						Expect(ui.Err).To(MatchError("Error: do not have any VPC security groups to associate with RDS instance"))
					})
				})
			})
		})

		Context("refresh", func() {
			var ui MockUi
			var conn *pluginfakes.FakeCliConnection
			var fakeRDSSvc *fakes.FakeRDSService
			var p *BasicPlugin
			var args []string

			BeforeEach(func() {
				conn = &pluginfakes.FakeCliConnection{}
				ui = MockUi{}
				fakeRDSSvc = &fakes.FakeRDSService{}

				p = &BasicPlugin{
					UI: &ui,
					Svc: fakeRDSSvc,
					WaitDuration: time.Millisecond,
				}
				args = []string{"aws-rds", "refresh", "name"}

				fakeRDSSvc.DescribeDBInstancesStub = func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
					if fakeRDSSvc.DescribeDBInstancesCallCount() < 5 {
						return &rds.DescribeDBInstancesOutput{
							DBInstances: []*rds.DBInstance{{
								DBInstanceIdentifier: aws.String("name"),
								DBInstanceStatus: aws.String("creating"),
							}},
						}, nil
					}

					return &rds.DescribeDBInstancesOutput{
						DBInstances: []*rds.DBInstance{{
							DBInstanceIdentifier: aws.String("name"),
							DBInstanceStatus: aws.String("available"),
							Endpoint: &rds.Endpoint {
								Port: aws.Int64(5432),
								Address: aws.String("test-uri.us-east-1.rds.amazonaws.com"),
							},
							DBInstanceArn: aws.String("arn:aws:rds:us-east-1:10101010:db:name"),
							DbiResourceId: aws.String("resourceid"),
							MasterUsername: aws.String("root"),
							DBName: aws.String("database"),
							VpcSecurityGroups: []*rds.VpcSecurityGroupMembership {{
								VpcSecurityGroupId: aws.String("vpcgroup"),
							}},
							DBSubnetGroup: &rds.DBSubnetGroup{
								DBSubnetGroupName: aws.String("default-vpc-vpcid"),
								VpcId: aws.String("vpcid"),
							},
						}},
					}, nil
				}

				fakeRDSSvc.ModifyDBInstanceReturns(&rds.ModifyDBInstanceOutput{}, nil)

				GenerateRandomAlphanumericString = func() string {
					return "password2"
				}
			})

			It("recaptures DB info from AWS and resets the password", func() {
				p.Run(conn, args)
				Expect(fakeRDSSvc.ModifyDBInstanceCallCount()).To(Equal(1))
				Expect(fakeRDSSvc.ModifyDBInstanceArgsForCall(0)).To(Equal(&rds.ModifyDBInstanceInput{
					DBInstanceIdentifier: aws.String("name"),
					MasterUserPassword: aws.String("password2"),
				}))
			})

			It("restarts the polling process", func() {
				p.Run(conn, args)
				Expect(fakeRDSSvc.DescribeDBInstancesCallCount()).To(Equal(6))
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
				Expect(ui.TextTemplate).To(Equal("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml"))
				Expect(ui.Data["Name"]).To(Equal("name"))
				Expect(ui.Data["RDSID"]).To(Equal("resourceid"))
				Expect(ui.Data["VPC"]).To(Equal("vpcid"))
				Expect(ui.Data["SecGroup"]).To(Equal("vpcgroup"))
			})

			Context("error cases", func() {
				Context("when describe DB instances returns no instances", func() {
					BeforeEach(func() {
						fakeRDSSvc.DescribeDBInstancesStub = func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
							return &rds.DescribeDBInstancesOutput{DBInstances: []*rds.DBInstance{}}, nil
						}
					})

					It("returns an error", func() {
						p.Run(conn, args)
						Expect(ui.Err).To(MatchError("Could not find db instance name"))
					})
				})
			})
		})
	})
})

// test case for error from cli command
type MockUi struct {
	TextTemplate string
	Err error
	Data map[string]interface{}
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