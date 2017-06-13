package cf_rds

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/jessevdk/go-flags"
	"github.com/seattle-beach/cf-cli-rds-plugin/api"
)

type UpsOption struct {
	Uri string `json:"uri"`
}

type TinyUI interface {
	DisplayError(err error)
	DisplayText(template string, data ...map[string]interface{})
	DisplayKeyValueTable(prefix string, table [][]string, padding int)
}

type Api interface {
	GetSubnetGroups() ([]*rds.DBSubnetGroup, error)
	CreateInstance(instance *api.DBInstance) (chan error, error)
	RefreshInstance(instance *api.DBInstance) chan error
}

type BasicPlugin struct {
	UI           TinyUI
	Api          Api
	WaitDuration time.Duration
}

func (c *BasicPlugin) createUPS(instance *api.DBInstance, cli plugin.CliConnection) error {
	serviceInfo, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	_, err = cli.CliCommand("cups", instance.InstanceName, "-p", string(serviceInfo))
	if err != nil {
		return err
	}

	return nil
}

func (c *BasicPlugin) updateUPS(instance *api.DBInstance, cli plugin.CliConnection) error {
	serviceInfo, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	_, err = cli.CliCommand("uups", instance.InstanceName, "-p", string(serviceInfo))
	if err != nil {
		return err
	}

	return nil
}

func (c *BasicPlugin) waitForApiResponse(instance *api.DBInstance, errChan chan error, cli plugin.CliConnection) {
	ticker := time.NewTicker(1 * c.WaitDuration).C

	for {
		select {
		case err := <-errChan:
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			err = c.createUPS(instance, cli)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			c.UI.DisplayText("AWS RDS Instance:\n{{.instance}}", map[string]interface{}{
				"instance": instance.InstanceName,
			})
			c.UI.DisplayKeyValueTable("", [][]string{
				{"ARN:", instance.ARN},
				{"RDSID:", instance.ResourceID},
				{"VPC:", *instance.SubnetGroup.VpcId},
				{"SecGroup:", *instance.SecGroups[0].VpcSecurityGroupId}},
				2)
			return
		case <- ticker:
			nextCheckTime := time.Now().Add(c.WaitDuration)
			c.UI.DisplayText("RDS instance not available yet, next check at {{.Time}}", map[string]interface{}{
				"Time": nextCheckTime.Format("15:04:05"),
			})
		}
	}
}

type AwsRdsOptions interface {
	SetServiceName(string)
}

func getOptions(opts AwsRdsOptions, cliConnection plugin.CliConnection, args []string) {
	parser := flags.NewParser(opts, flags.None)
	extraArgs, err := parser.ParseArgs(args[1:])

	handleErrors(args[0], err, extraArgs, cliConnection)

	opts.SetServiceName(extraArgs[0])
}

func handleErrors(cmd string, err error, args []string, cliConnection plugin.CliConnection) {
	if err != nil {
		fmt.Println(fmt.Sprintf("Incorrect Usage: %v", err))
		cliConnection.CliCommand("help", cmd)
		os.Exit(0)
	}

	if len(args) != 1 {
		cliConnection.CliCommand("help", cmd)
		os.Exit(0)
	}
}

type AwsRdsCreateOptions struct {
	ServiceName string
	Engine      string `long:"engine" description:"The name of the RDS database engine to be used for this instance." required:"false" default:"postgres"`
	Storage     int64  `long:"size" description:"The amount of storage in Gb for the RDS instance." required:"false" default:"20"`
	Class       string `long:"class" description:"The RDS instance type class." required:"false" default:"db.t2.micro"`
}

func (a *AwsRdsCreateOptions) SetServiceName(name string) {
	a.ServiceName = name
}

func (c *BasicPlugin) AwsRdsCreateRun(cliConnection plugin.CliConnection, args []string) {
	opts := AwsRdsCreateOptions{}
	getOptions(&opts, cliConnection, args)

	subnetGroups, err := c.Api.GetSubnetGroups()
	if err != nil {
		c.UI.DisplayError(err)
		return
	}

	dbInstance := &api.DBInstance{
		InstanceName:  opts.ServiceName,
		SubnetGroup:   subnetGroups[0],
		InstanceClass: opts.Class,
		Engine:        opts.Engine,
		Storage:       opts.Storage,
		AZ:            "us-east-1a",
		Port:          int64(5432),
		Username:      "root",
	}

	errChan, err := c.Api.CreateInstance(dbInstance)
	if err != nil {
		c.UI.DisplayError(err)
		return
	}
	c.UI.DisplayText("Creating RDS Instance. This may take several minutes...")
	c.waitForApiResponse(dbInstance, errChan, cliConnection)
}

type AwsRdsRefreshOptions struct {
	ServiceName string
}

func (a *AwsRdsRefreshOptions) SetServiceName(name string) {
	a.ServiceName = name
}

func (c *BasicPlugin) AwsRdsRefreshRun(cliConnection plugin.CliConnection, args []string) {
	opts := AwsRdsRefreshOptions{}
	getOptions(&opts, cliConnection, args)

	dbInstance := &api.DBInstance{
		InstanceName: opts.ServiceName,
	}

	errChan := c.Api.RefreshInstance(dbInstance)
	c.waitForApiResponse(dbInstance, errChan, cliConnection)
}

type AwsRdsRegisterOptions struct {
	ServiceName string
	Uri         string `long:"uri" description:"" required:"true"`
}

func (a *AwsRdsRegisterOptions) SetServiceName(name string) {
	a.ServiceName = name
}

func (c *BasicPlugin) AwsRdsRegisterRun(cliConnection plugin.CliConnection, args []string) {
	opts := AwsRdsRegisterOptions{}
	getOptions(&opts, cliConnection, args)

	uri, _ := json.Marshal(&UpsOption{
		Uri: opts.Uri,
	})
	_, err := cliConnection.CliCommand("cups", opts.ServiceName, "-p", string(uri))
	if err != nil {
		c.UI.DisplayError(err)
		return
	}

	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		c.UI.DisplayError(err)
		return
	}

	c.UI.DisplayText("Successfully created user-provided service {{.ServiceName}} in space {{.Space}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml",
		map[string]interface{}{
			"ServiceName": opts.ServiceName,
			"Space":       space.Name,
		},
	)
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "aws-rds-create":
		c.AwsRdsCreateRun(cliConnection, args)
		return
	case "aws-rds-refresh":
		c.AwsRdsRefreshRun(cliConnection, args)
		return
	case "aws-rds-register":
		c.AwsRdsRegisterRun(cliConnection, args)
		return
	default:
		// TODO Show Usage
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
					Usage: "cf aws-rds-create [--engine ENGINE] [--size SIZE] [--class CLASS] SERVICE_NAME",
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
	}
}
