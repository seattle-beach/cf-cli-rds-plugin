package cf_rds

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
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
}

type RDSService interface {
	DescribeDBSubnetGroups(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error)
	CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error)
}

// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second parameter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
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

			subnetGroupName := *subnetGroupsResp.DBSubnetGroups[0].DBSubnetGroupName
			params := &rds.CreateDBInstanceInput{
				DBInstanceClass:         aws.String("db.t2.micro"), // Required
				DBInstanceIdentifier:    aws.String(args[2]), // Required
				Engine:                  aws.String("postgres"), // Required
				AllocatedStorage:        aws.Int64(20),
				AutoMinorVersionUpgrade: aws.Bool(true),
				AvailabilityZone:        aws.String("us-east-1a"),
				CopyTagsToSnapshot:      aws.Bool(true),
				DBName:                  aws.String(fmt.Sprintf("%sdb", args[2])),
				DBParameterGroupName:    aws.String("default.postgres9.6"),
				DBSubnetGroupName:       aws.String(subnetGroupName),
				MasterUserPassword:      aws.String("password"),
				MasterUsername:          aws.String("root"),
				MultiAZ:                 aws.Bool(false),
				Port:                    aws.Int64(5432),
				PubliclyAccessible:      aws.Bool(true),
			}

			createDBInstanceResp, err := c.Svc.CreateDBInstance(params)

			if err != nil {
				fmt.Println(err.Error())
				return
			}

			_, err = cliConnection.CliCommand("cups", args[2])
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			resourceID := createDBInstanceResp.DBInstance.DbiResourceId
			secGroup := createDBInstanceResp.DBInstance.VpcSecurityGroups[0].VpcSecurityGroupId

			c.UI.DisplayText("Successfully created user-provided service {{.Name}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml", map[string]interface{}{
				"Name": args[2],
				"RDSID": *resourceID,
				"VPC": *subnetGroupsResp.DBSubnetGroups[0].VpcId,
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


// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as separate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.

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
