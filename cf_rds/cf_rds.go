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

// BasicPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type BasicPlugin struct {
	UI  TinyUI
	Svc RDSService
	WaitDuration time.Duration
}

type RDSService interface {
	DescribeDBSubnetGroups(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error)
	CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
}

type DBInstance struct {
	ARN string `json:"arn"`
	ResourceID string `json:"resource_id"`
	DBURI string `json:"uri,omitempty"`
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[0] == "aws-rds" {
		if len(args) > 2 && args[1] == "create" {
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
				c.UI.DisplayError(err)
				return
			}

			subnetGroups := subnetGroupsResp.DBSubnetGroups
			if len(subnetGroups) == 0 {
				c.UI.DisplayError(errors.New("Error: did not find any DB subnet groups to create RDS instance in"))
				return
			}

			subnetGroupName := *subnetGroups[0].DBSubnetGroupName
			dbName := GenerateRandomString()
			dbPassword := GenerateRandomAlphanumericString()

			createDBInstanceResp, err := c.Svc.CreateDBInstance(&rds.CreateDBInstanceInput{
				DBInstanceClass:         aws.String("db.t2.micro"), // Required
				DBInstanceIdentifier:    aws.String(args[2]),       // Required
				Engine:                  aws.String("postgres"),    // Required
				AllocatedStorage:        aws.Int64(20),
				AutoMinorVersionUpgrade: aws.Bool(true),
				AvailabilityZone:        aws.String("us-east-1a"),
				CopyTagsToSnapshot:      aws.Bool(true),
				DBName:                  aws.String(dbName),
				DBParameterGroupName:    aws.String("default.postgres9.6"),
				DBSubnetGroupName:       aws.String(subnetGroupName),
				MasterUserPassword:      aws.String(dbPassword),
				MasterUsername:          aws.String("root"),
				MultiAZ:                 aws.Bool(false),
				Port:                    aws.Int64(5432),
				PubliclyAccessible:      aws.Bool(true),
			})
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			resourceID := createDBInstanceResp.DBInstance.DbiResourceId
			arn := createDBInstanceResp.DBInstance.DBInstanceArn
			vpcSecGroups := createDBInstanceResp.DBInstance.VpcSecurityGroups

			if len(vpcSecGroups) == 0 {
				c.UI.DisplayError(errors.New("Error: do not have any VPC security groups to associate with RDS instance"))
				return
			}
			secGroup := vpcSecGroups[0].VpcSecurityGroupId

			dbI := DBInstance{
				ResourceID: *resourceID,
				ARN: *arn,
			}
			serviceInfo, err := json.Marshal(&dbI)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			_, err = cliConnection.CliCommand("cups", args[2], "-p", string(serviceInfo))
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			var dbAddr *string
			var dbPort *int64

			for {
				describeDBInstancesResp, err := c.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: aws.String(args[2]),
				})
				if err != nil {
					c.UI.DisplayError(err)
					return
				}

				dbInstanceStatus := describeDBInstancesResp.DBInstances[0].DBInstanceStatus
				if *dbInstanceStatus == "available" {
					dbAddr = describeDBInstancesResp.DBInstances[0].Endpoint.Address
					dbPort = describeDBInstancesResp.DBInstances[0].Endpoint.Port
					break
				}

				nextCheckTime := time.Now().Add(c.WaitDuration)
				c.UI.DisplayText("Checking connectivity... not available yet, will check again at {{.Time}}", map[string]interface{}{
					"Time": nextCheckTime.Format("15:04:05"),
				})
				time.Sleep(1 * c.WaitDuration)
			}

			dbI = DBInstance{
				ResourceID: *resourceID,
				ARN: *arn,
				DBURI: fmt.Sprintf("postgres://root:%s@%s:%d/%s", dbPassword, *dbAddr, *dbPort, dbName),
			}

			serviceInfo, err = json.Marshal(&dbI)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			_, err = cliConnection.CliCommand("uups", args[2], "-p", string(serviceInfo))

			c.UI.DisplayText("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml", map[string]interface{}{
				"Name":     args[2],
				"RDSID":    *resourceID,
				"VPC":      *subnetGroups[0].VpcId,
				"SecGroup": *secGroup,
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

		c.UI.DisplayError(errors.New(fmt.Sprintf("%s\n%s", "Usage: cf aws-rds register NAME --uri URI",
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