package cf_rds

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/errors"
	"time"
	"github.com/seattle-beach/cf-cli-rds-plugin/api"
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
	Api *api.CfRDSApi
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
	for {
		select {
		case err := <-errChan:
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			err = c.updateUPS(instance, cli)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			c.UI.DisplayText("Successfully created user-provided service {{.ServiceName}} exposing RDS Instance {{.Name}}, {{.RDSID}} in AWS VPC {{.VPC}} with Security Group {{.SecGroup}}! You can bind this service to an app using `cf bind-service` or add it to the `services` section in your manifest.yml", map[string]interface{}{
				"ServiceName":     instance.InstanceName,
				"RDSID":    instance.ResourceID,
				"VPC":      *instance.SubnetGroup.VpcId,
				"SecGroup": *instance.SecGroups[0].VpcSecurityGroupId,
			})
			return
		default:
			nextCheckTime := time.Now().Add(c.WaitDuration)
			c.UI.DisplayText("Checking connectivity... not available yet, will check again at {{.Time}}", map[string]interface{}{
				"Time": nextCheckTime.Format("15:04:05"),
			})
			time.Sleep(1 * c.WaitDuration)
		}
	}
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
		if args[0] == "aws-rds-create" && len(args) > 1 {
			serviceName := args[1]
			subnetGroups, err := c.Api.GetSubnetGroups()
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			dbInstance := &api.DBInstance{
				InstanceName: serviceName,
				SubnetGroup: subnetGroups[0],
				InstanceClass: "db.t2.micro",
				Engine: "postgres",
				Storage: int64(20),
				AZ: "us-east-1a",
				Port: int64(5432),
				Username: "root",
			}

			err = c.createUPS(dbInstance, cliConnection)
			if err != nil {
				c.UI.DisplayError(err)
				return
			}

			errChan := c.Api.CreateInstance(dbInstance)
			c.waitForApiResponse(dbInstance, errChan, cliConnection)
			return
		}

		if args[0] == "aws-rds-refresh" && len(args) > 1 {
			serviceName := args[1]
			dbInstance := &api.DBInstance{
				InstanceName: serviceName,
			}

			errChan := c.Api.RefreshInstance(dbInstance)
			c.waitForApiResponse(dbInstance, errChan, cliConnection)
			return
		}

		if args[0] == "aws-rds-register" && len(args) == 4 && args[2] == "--uri" {
			serviceName := args[1]
			uri, _ := json.Marshal(&UpsOption{
				Uri: args[3],
			})
			_, err := cliConnection.CliCommand("cups", serviceName, "-p", string(uri))
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
					"ServiceName":  serviceName,
					"Space": space.Name,
				},
			)
			return
		}

		switch args[0] {
		case "aws-rds-register" :
			c.UI.DisplayError(errors.New(c.GetMetadata().Commands[0].UsageDetails.Usage))
		case "aws-rds-create" :
			c.UI.DisplayError(errors.New(c.GetMetadata().Commands[1].UsageDetails.Usage))
		case "aws-rds-refresh":
			c.UI.DisplayError(errors.New(c.GetMetadata().Commands[2].UsageDetails.Usage))
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
	}
}