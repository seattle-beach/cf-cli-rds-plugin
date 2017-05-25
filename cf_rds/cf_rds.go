package cf_rds

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"time"
	"math/rand"
)

type UpsOption struct {
	Uri string `json:"uri"`
}

type TinyUI interface {
	DisplayError(err error)
	DisplayText(template string, data ...map[string]interface{})
}

type BasicPlugin struct {
	UI  TinyUI
	Svc RDSService
	WaitDuration time.Duration
}

type RDSService interface {
	DescribeDBSubnetGroups(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error)
	CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
	ModifyDBInstance(input *rds.ModifyDBInstanceInput) (*rds.ModifyDBInstanceOutput, error)
}

type DBInstance struct {
	ARN string `json:"arn,omitempty"`
	InstanceName string `json:"instance_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	DBURI string `json:"uri,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	DBName string `json:"database,omitempty"`
	SecGroups []*rds.VpcSecurityGroupMembership `json:"-"`
	VPCID string `json:"-"`
}

func (c *BasicPlugin) getSubnetGroups() ([]*rds.DBSubnetGroup, error) {
	subnetGroupsResp, err := c.Svc.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{
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
	})
	if err != nil {
		return []*rds.DBSubnetGroup{}, err
	}

	subnetGroups := subnetGroupsResp.DBSubnetGroups
	if len(subnetGroups) == 0 {
		return subnetGroups, errors.New("Error: did not find any DB subnet groups to create RDS instance in")
	}

	return subnetGroups, nil
}

func (c *BasicPlugin) createRDSInstance(instanceName string, subnetGroup *rds.DBSubnetGroup) (DBInstance, error) {
	dbName := GenerateRandomString()
	dbPassword := GenerateRandomAlphanumericString()

	createDBInstanceResp, err := c.Svc.CreateDBInstance(&rds.CreateDBInstanceInput{
		DBInstanceClass:         aws.String("db.t2.micro"), // Required
		DBInstanceIdentifier:    aws.String(instanceName),       // Required
		Engine:                  aws.String("postgres"),    // Required
		AllocatedStorage:        aws.Int64(20),
		AutoMinorVersionUpgrade: aws.Bool(true),
		AvailabilityZone:        aws.String("us-east-1a"),
		CopyTagsToSnapshot:      aws.Bool(true),
		DBName:                  aws.String(dbName),
		DBParameterGroupName:    aws.String("default.postgres9.6"),
		DBSubnetGroupName:       subnetGroup.DBSubnetGroupName,
		MasterUserPassword:      aws.String(dbPassword),
		MasterUsername:          aws.String("root"),
		MultiAZ:                 aws.Bool(false),
		Port:                    aws.Int64(5432),
		PubliclyAccessible:      aws.Bool(true),
	})
	if err != nil {
		return DBInstance{}, err
	}

	secGroups := createDBInstanceResp.DBInstance.VpcSecurityGroups
	if len(secGroups) == 0 {
		return DBInstance{}, errors.New("Error: do not have any VPC security groups to associate with RDS instance")
	}

	return DBInstance{
		InstanceName: instanceName,
		ARN: *createDBInstanceResp.DBInstance.DBInstanceArn,
		ResourceID: *createDBInstanceResp.DBInstance.DbiResourceId,
		Username: "root",
		Password: dbPassword,
		DBName: dbName,
		SecGroups: secGroups,
		VPCID: *subnetGroup.VpcId,
	}, nil
}

func (c *BasicPlugin) refreshRDSInstanceInfo(instance *DBInstance) error {
	describeDBInstancesResp, err := c.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
	})
	if err != nil {
		return err
	}

	dbInstances := describeDBInstancesResp.DBInstances
	newPassword := GenerateRandomAlphanumericString()
	_, err = c.Svc.ModifyDBInstance(&rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
		MasterUserPassword: aws.String(newPassword),
	})
	if err != nil {
		return err
	}

	instance.ARN = *dbInstances[0].DBInstanceArn
	instance.ResourceID = *dbInstances[0].DbiResourceId
	instance.Username = *dbInstances[0].MasterUsername
	instance.DBName = *dbInstances[0].DBName
	instance.Password = newPassword
	instance.SecGroups = dbInstances[0].VpcSecurityGroups
	instance.VPCID = *dbInstances[0].DBSubnetGroup.VpcId

	dbAddr := dbInstances[0].Endpoint.Address
	dbPort := dbInstances[0].Endpoint.Port
	instance.DBURI = fmt.Sprintf("postgres://root:%s@%s:%d/%s", instance.Password, *dbAddr, *dbPort, instance.DBName)

	return nil
}

func (c *BasicPlugin) populateURIAfterCreate(instance *DBInstance) error {
	describeDBInstancesResp, err := c.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
	})
	if err != nil {
		return err
	}

	dbInstances := describeDBInstancesResp.DBInstances
	dbAddr := dbInstances[0].Endpoint.Address
	dbPort := dbInstances[0].Endpoint.Port
	instance.DBURI = fmt.Sprintf("postgres://root:%s@%s:%d/%s", instance.Password, *dbAddr, *dbPort, instance.DBName)

	return nil
}

func (c *BasicPlugin) pollUntilDBAvailable(instance *DBInstance) error {
	for {
		describeDBInstancesResp, err := c.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(instance.InstanceName),
		})
		if err != nil {
			return err
		}

		dbInstances := describeDBInstancesResp.DBInstances
		if len(dbInstances) == 0 {
			return fmt.Errorf("Could not find db instance %s", instance.InstanceName)
		}

		dbInstanceStatus := dbInstances[0].DBInstanceStatus
		if *dbInstanceStatus == "available" {
			break
		}

		nextCheckTime := time.Now().Add(c.WaitDuration)
		c.UI.DisplayText("Checking connectivity... not available yet, will check again at {{.Time}}", map[string]interface{}{
			"Time": nextCheckTime.Format("15:04:05"),
		})
		time.Sleep(1 * c.WaitDuration)
	}

	return nil
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "aws-rds" {
		if len(args) > 2 && args[1] == "create" {
			serviceName := args[2]
			subnetGroups, err := c.getSubnetGroups()
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			dbInstance, err := c.createRDSInstance(serviceName, subnetGroups[0])
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			serviceInfo, err := json.Marshal(&dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			_, err = cliConnection.CliCommand("cups", serviceName, "-p", string(serviceInfo))
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			err = c.pollUntilDBAvailable(&dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			err = c.populateURIAfterCreate(&dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			serviceInfo, err = json.Marshal(&dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			_, err = cliConnection.CliCommand("uups", serviceName, "-p", string(serviceInfo))
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			c.UI.DisplayText("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml", map[string]interface{}{
				"Name":     serviceName,
				"RDSID":    dbInstance.ResourceID,
				"VPC":      dbInstance.VPCID,
				"SecGroup": *dbInstance.SecGroups[0].VpcSecurityGroupId,
			})
			return
		}

		if len(args) > 2 && args[1] == "refresh" {
			name := args[2]
			dbInstance := &DBInstance{}
			dbInstance.InstanceName = name

			err := c.pollUntilDBAvailable(dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			err = c.refreshRDSInstanceInfo(dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			serviceInfo, err := json.Marshal(dbInstance)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			_, err = cliConnection.CliCommand("uups", name, "-p", string(serviceInfo))
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			c.UI.DisplayText("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml", map[string]interface{}{
				"Name":     name,
				"RDSID":    dbInstance.ResourceID,
				"VPC":      dbInstance.VPCID,
				"SecGroup": *dbInstance.SecGroups[0].VpcSecurityGroupId,
			})
			return
		}

		if len(args) == 5 && args[1] == "register" && args[3] == "--uri" {
			name := args[2]
			uri, _ := json.Marshal(&UpsOption{
				Uri: args[4],
			})
			_, err := cliConnection.CliCommand("cups", name, "-p", string(uri))
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			space, err := cliConnection.GetCurrentSpace()
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			c.UI.DisplayText("Successfully created user-provided service {{.Name}} in space {{.Space}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml",
				map[string]interface{}{
					"Name":  name,
					"Space": space.Name,
				},
			)
			return
		}

		c.UI.DisplayError(errors.New(fmt.Sprintf("Usage:\n%s\n%s", "cf aws-rds register NAME --uri URI",
			"cf aws-rds create NAME")))
		return
	}
}

func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
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
				Name:     "aws-rds",
				HelpText: "plugin to hook up rds to pws",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "aws-rds\n   cf aws-rds",
				},
			},
		},
	}
}

var GenerateRandomString = func() string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

var GenerateRandomAlphanumericString = func() string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}