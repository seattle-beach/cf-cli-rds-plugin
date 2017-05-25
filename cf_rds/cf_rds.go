package cf_rds

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/errors"
	"fmt"
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
	if args[0] == "aws-rds" {
		if len(args) > 2 && args[1] == "create" {
			serviceName := args[2]
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

		if len(args) > 2 && args[1] == "refresh" {
			serviceName := args[2]
			dbInstance := &api.DBInstance{
				InstanceName: serviceName,
			}

			errChan := c.Api.RefreshInstance(dbInstance)
			c.waitForApiResponse(dbInstance, errChan, cliConnection)
			return
		}

		if len(args) == 5 && args[1] == "register" && args[3] == "--uri" {
			serviceName := args[2]
			uri, _ := json.Marshal(&UpsOption{
				Uri: args[4],
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

		c.UI.DisplayError(errors.New(fmt.Sprintf("Usage:\n%s\n%s\n%s",
			"cf aws-rds register SERVICE_NAME --uri URI",
			"cf aws-rds create SERVICE_NAME",
			"cf aws-rds refresh SERVICE_NAME",
		)))
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

				UsageDetails: plugin.Usage{
					Usage: "aws-rds\n   cf aws-rds",
				},
			},
		},
	}
}