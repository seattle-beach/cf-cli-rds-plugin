package api_test

import (
	."github.com/onsi/ginkgo"
	."github.com/onsi/gomega"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/seattle-beach/cf-cli-rds-plugin/api/fakes"
	"github.com/seattle-beach/cf-cli-rds-plugin/api"
	"errors"
)

var _ = Describe("Api", func() {
	var fakeRDSSvc *fakes.FakeRDSService
	var cfRDSApi *api.CfRDSApi

	BeforeEach(func() {
		fakeRDSSvc = &fakes.FakeRDSService{}
		cfRDSApi = &api.CfRDSApi{
			Svc: fakeRDSSvc,
		}
	})

	Describe("GetSubnetGroups", func() {
		BeforeEach(func() {
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
		})

		It("returns a list of DB subnet groups", func() {
			subnetGroups, err := cfRDSApi.GetSubnetGroups()
			Expect(err).NotTo(HaveOccurred())
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
			Expect(subnetGroups).To(Equal([]*rds.DBSubnetGroup{{
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
			}}))
		})

		Context("Error cases", func() {
			Context("when no AWS credentials are provided", func() {
				BeforeEach(func() {
					fakeRDSSvc.DescribeDBSubnetGroupsReturns(&rds.DescribeDBSubnetGroupsOutput{
						DBSubnetGroups: []*rds.DBSubnetGroup{},
					}, errors.New("NoCredentialProviders"))
				})

				It("should return helpful error", func() {
					_, err := cfRDSApi.GetSubnetGroups()
					Expect(err).To(MatchError("No valid AWS credentials found. Please see this document for help configuring the AWS SDK: https://github.com/aws/aws-sdk-go#configuring-credentials"))
				})
			})

			Context("when there are no DB subnet groups", func() {
				BeforeEach(func() {
					fakeRDSSvc.DescribeDBSubnetGroupsReturns(&rds.DescribeDBSubnetGroupsOutput{
						DBSubnetGroups: []*rds.DBSubnetGroup{},
					}, nil)
				})

				It("should return an error", func() {
					_, err := cfRDSApi.GetSubnetGroups()
					Expect(err).To(MatchError("Error: did not find any DB subnet groups to create RDS instance in"))
				})
			})
		})
	})

	Describe("CreateInstance", func() {
		var instance *api.DBInstance

		BeforeEach(func() {
			instance = &api.DBInstance{
				InstanceName: "test-instance",
				SubnetGroup: &rds.DBSubnetGroup{
					DBSubnetGroupName: aws.String("default-vpc-vpcid"),
				},
				InstanceClass: "db.t1.micro",
				Engine: "postgres",
				Storage: int64(20),
				AZ: "us-east-1a",
				Port: int64(5432),
				Username: "root",
			}

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
				return &rds.DescribeDBInstancesOutput{
					DBInstances: []*rds.DBInstance{{
						DBInstanceIdentifier: aws.String("test-instance"),
						DBInstanceStatus: aws.String("available"),
						Endpoint: &rds.Endpoint {
							Port: aws.Int64(5432),
							Address: aws.String("test-uri.us-east-1.rds.amazonaws.com"),
						},
					}},
				}, nil
			}

			fakeRDSSvc.WaitUntilDBInstanceAvailableReturns(nil)

			api.GenerateRandomAlphanumericString = func() string {
				return "password"
			}

			api.GenerateRandomString = func() string {
				return "database"
			}
		})

		It("creates an RDS instance with the given parameters", func() {
			errChan, err := cfRDSApi.CreateInstance(instance)
			Expect(err).NotTo(HaveOccurred())
			err = <- errChan
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRDSSvc.CreateDBInstanceCallCount()).To(Equal(1))

			createDBInstanceInput := fakeRDSSvc.CreateDBInstanceArgsForCall(0)
			Expect(createDBInstanceInput.DBInstanceClass).To(Equal(aws.String("db.t1.micro")))
			Expect(createDBInstanceInput.DBInstanceIdentifier).To(Equal(aws.String("test-instance")))
			Expect(createDBInstanceInput.Engine).To(Equal(aws.String("postgres")))
			Expect(createDBInstanceInput.AllocatedStorage).To(Equal(aws.Int64(20)))
			Expect(createDBInstanceInput.AvailabilityZone).To(Equal(aws.String("us-east-1a")))
			Expect(createDBInstanceInput.DBSubnetGroupName).To(Equal(aws.String("default-vpc-vpcid")))
			Expect(createDBInstanceInput.Port).To(Equal(aws.Int64(5432)))

			Expect(instance.ARN).To(Equal("arn:aws:rds:us-east-1:10101010:db:name"))
			Expect(instance.ResourceID).To(Equal("resourceid"))
			Expect(instance.SecGroups).To(Equal([]*rds.VpcSecurityGroupMembership{{
				VpcSecurityGroupId: aws.String("vpcgroup"),
			}}))
			Expect(instance.Username).To(Equal("root"))
		})

		It("polls until the RDS instance is available", func() {
			errChan, _ := cfRDSApi.CreateInstance(instance)
			err := <- errChan
			Expect(err).NotTo(HaveOccurred())

			Expect(instance.DBURI).To(Equal("postgres://root:password@test-uri.us-east-1.rds.amazonaws.com:5432/database"))

			Expect(fakeRDSSvc.WaitUntilDBInstanceAvailableCallCount()).To(Equal(1))
		})

		Context("error cases", func() {
			Context("when there are no vpc security groups", func() {
				BeforeEach(func() {
					fakeRDSSvc.CreateDBInstanceReturns(&rds.CreateDBInstanceOutput{
						DBInstance: &rds.DBInstance{
							DbiResourceId:     aws.String("resourceid"),
							DBInstanceArn:     aws.String("resourcearn"),
							VpcSecurityGroups: []*rds.VpcSecurityGroupMembership{},
						},
					}, nil)
				})

				It("should return an error", func() {
					_, err := cfRDSApi.CreateInstance(instance)
					Expect(err).To(MatchError("Error: do not have any VPC security groups to associate with RDS instance"))
				})
			})
		})
	})

	Describe("RefreshInstance", func() {
		var instance *api.DBInstance

		BeforeEach(func() {
			instance = &api.DBInstance{
				InstanceName: "test-instance",
			}

			fakeRDSSvc.DescribeDBInstancesStub = func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
				return &rds.DescribeDBInstancesOutput{
					DBInstances: []*rds.DBInstance{{
						DBInstanceIdentifier: aws.String("test-instance"),
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
						Engine: aws.String("postgres"),
					}},
				}, nil
			}

			fakeRDSSvc.WaitUntilDBInstanceAvailableReturns(nil)

			api.GenerateRandomAlphanumericString = func() string {
				return "password_reset"
			}
		})

		It("updates the given instance object with information when it is available", func() {
			errChan := cfRDSApi.RefreshInstance(instance)
			err := <-errChan
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRDSSvc.WaitUntilDBInstanceAvailableCallCount()).To(Equal(1))
			Expect(instance.ARN).To(Equal("arn:aws:rds:us-east-1:10101010:db:name"))
			Expect(instance.ResourceID).To(Equal("resourceid"))
			Expect(instance.SecGroups).To(Equal([]*rds.VpcSecurityGroupMembership{{
				VpcSecurityGroupId: aws.String("vpcgroup"),
			}}))
			Expect(instance.Username).To(Equal("root"))
			Expect(instance.DBName).To(Equal("database"))
			Expect(instance.SubnetGroup).To(Equal(&rds.DBSubnetGroup{
					DBSubnetGroupName: aws.String("default-vpc-vpcid"),
					VpcId: aws.String("vpcid"),
			}))
			Expect(instance.Engine).To(Equal("postgres"))
		})

		It("modifies the instance master user password", func() {
			errChan := cfRDSApi.RefreshInstance(instance)
			err := <-errChan
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRDSSvc.ModifyDBInstanceCallCount()).To(Equal(1))

			Expect(fakeRDSSvc.ModifyDBInstanceArgsForCall(0)).To(Equal(&rds.ModifyDBInstanceInput{
				DBInstanceIdentifier: aws.String("test-instance"),
				MasterUserPassword: aws.String("password_reset"),
			}))
			Expect(instance.DBURI).To(Equal("postgres://root:password_reset@test-uri.us-east-1.rds.amazonaws.com:5432/database"))
		})

		Context("error cases", func() {
			Context("when no AWS credentials are provided", func() {
				BeforeEach(func() {
					fakeRDSSvc.DescribeDBInstancesReturns(&rds.DescribeDBInstancesOutput{}, errors.New("NoCredentialProviders"))
				})

				It("should return helpful error", func() {
					errChan := cfRDSApi.RefreshInstance(instance)
					err := <-errChan

					Expect(err).To(MatchError("No valid AWS credentials found. Please see this document for help configuring the AWS SDK: https://github.com/aws/aws-sdk-go#configuring-credentials"))
				})
			})

			Context("when describe DB instances returns no instances", func() {
				BeforeEach(func() {
					fakeRDSSvc.DescribeDBInstancesStub = func(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
						return &rds.DescribeDBInstancesOutput{DBInstances: []*rds.DBInstance{}}, nil
					}
				})

				It("returns an error", func() {
					errChan := cfRDSApi.RefreshInstance(instance)
					err := <-errChan

					Expect(err).To(MatchError("Could not find db instance test-instance"))
				})
			})
		})
	})
})