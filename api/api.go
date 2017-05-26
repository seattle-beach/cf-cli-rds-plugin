package api

import (
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
	"time"
	"math/rand"
	"fmt"
	"strings"
)

type RDSService interface {
	DescribeDBSubnetGroups(input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error)
	CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
	ModifyDBInstance(input *rds.ModifyDBInstanceInput) (*rds.ModifyDBInstanceOutput, error)
	WaitUntilDBInstanceAvailable(input *rds.DescribeDBInstancesInput) error
}


type CfRDSApi struct {
	Svc RDSService
}

type DBInstance struct {
	ARN string `json:"arn,omitempty"`
	InstanceName string `json:"instance_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	DBName string `json:"database,omitempty"`
	DBURI string `json:"uri,omitempty"`

	SecGroups []*rds.VpcSecurityGroupMembership `json:"-"`
	SubnetGroup *rds.DBSubnetGroup `json:"-"`
	Engine string `json:"-"`
	InstanceClass string `json:"-"`
	Storage int64 `json:"-"`
	AZ string `json:"-"`
	Port int64 `json:"-"`
}

func (f *CfRDSApi) GetSubnetGroups() ([]*rds.DBSubnetGroup, error) {
	subnetGroupsResp, err := f.Svc.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{
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
		if strings.Contains(err.Error(), "NoCredentialProviders") {
			return []*rds.DBSubnetGroup{}, errors.New("No valid AWS credentials found. Please see this document for help configuring the AWS SDK: https://github.com/aws/aws-sdk-go#configuring-credentials")
		}
		return []*rds.DBSubnetGroup{}, err
	}

	subnetGroups := subnetGroupsResp.DBSubnetGroups
	if len(subnetGroups) == 0 {
		return subnetGroups, errors.New("Error: did not find any DB subnet groups to create RDS instance in")
	}

	return subnetGroups, nil
}

func (f *CfRDSApi) CreateInstance(instance *DBInstance) chan error {
	dbName := GenerateRandomString()
	dbPassword := GenerateRandomAlphanumericString()
	errChan := make(chan error, 1)

	var paramGroup string
	switch instance.Engine {
	case "postgres":
		paramGroup = "default.postgres9.6"
	default:
		paramGroup = ""
	}

	createDBInstanceResp, err := f.Svc.CreateDBInstance(&rds.CreateDBInstanceInput{
		DBInstanceClass:         aws.String(instance.InstanceClass),
		DBInstanceIdentifier:    aws.String(instance.InstanceName),
		Engine:                  aws.String(instance.Engine),
		AllocatedStorage:        aws.Int64(instance.Storage),
		AutoMinorVersionUpgrade: aws.Bool(true),
		AvailabilityZone:        aws.String(instance.AZ),
		CopyTagsToSnapshot:      aws.Bool(true),
		DBName:                  aws.String(dbName),
		DBParameterGroupName:    aws.String(paramGroup),
		DBSubnetGroupName:       instance.SubnetGroup.DBSubnetGroupName,
		MasterUserPassword:      aws.String(dbPassword),
		MasterUsername:          aws.String(instance.Username),
		MultiAZ:                 aws.Bool(false),
		Port:                    aws.Int64(instance.Port),
		PubliclyAccessible:      aws.Bool(true),
	})
	if err != nil {
		errChan <- err
		return errChan
	}

	secGroups := createDBInstanceResp.DBInstance.VpcSecurityGroups
	if len(secGroups) == 0 {
		errChan <- errors.New("Error: do not have any VPC security groups to associate with RDS instance")
		return errChan
	}

	instance.ARN = *createDBInstanceResp.DBInstance.DBInstanceArn
	instance.ResourceID = *createDBInstanceResp.DBInstance.DbiResourceId
	instance.SecGroups = createDBInstanceResp.DBInstance.VpcSecurityGroups
	instance.Password = dbPassword
	instance.DBName = dbName

	go f.waitForInstance(instance, errChan, false)

	return errChan
}

func (f *CfRDSApi) RefreshInstance(instance *DBInstance) chan error {
	errChan := make(chan error, 1)
	describeDBInstancesResp, err := f.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoCredentialProviders") {
			err = errors.New("No valid AWS credentials found. Please see this document for help configuring the AWS SDK: https://github.com/aws/aws-sdk-go#configuring-credentials")
		}

		errChan <- err
		return errChan
	}

	dbInstances := describeDBInstancesResp.DBInstances
	if len(dbInstances) == 0 {
		errChan <- fmt.Errorf("Could not find db instance %s", instance.InstanceName)
		return errChan
	}

	instance.ARN = *dbInstances[0].DBInstanceArn
	instance.ResourceID = *dbInstances[0].DbiResourceId
	instance.Username = *dbInstances[0].MasterUsername
	instance.DBName = *dbInstances[0].DBName
	instance.SecGroups = dbInstances[0].VpcSecurityGroups
	instance.SubnetGroup = dbInstances[0].DBSubnetGroup
	instance.Engine = *dbInstances[0].Engine

	go f.waitForInstance(instance, errChan, true)

	return errChan
}

func (f *CfRDSApi) waitForInstance(instance *DBInstance, errChan chan error, generateNewPassword bool) {
	err := f.Svc.WaitUntilDBInstanceAvailable(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
	})
	if err != nil {
		errChan <- err
		return
	}

	describeDBInstancesResp, err := f.Svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance.InstanceName),
	})
	if err != nil {
		errChan <- err
		return
	}

	dbInstances := describeDBInstancesResp.DBInstances
	if len(dbInstances) == 0 {
		errChan <- fmt.Errorf("Could not find db instance %s", instance.InstanceName)
		return
	}

	dbInstanceStatus := dbInstances[0].DBInstanceStatus
	if *dbInstanceStatus == "available" {
		if generateNewPassword {
			instance.Password = GenerateRandomAlphanumericString()
			_, err = f.Svc.ModifyDBInstance(&rds.ModifyDBInstanceInput{
				DBInstanceIdentifier: aws.String(instance.InstanceName),
				MasterUserPassword: aws.String(instance.Password),
			})
			if err != nil {
				errChan <- err
				return
			}
		}
		dbAddr := dbInstances[0].Endpoint.Address
		dbPort := dbInstances[0].Endpoint.Port
		instance.DBURI = fmt.Sprintf("%s://%s:%s@%s:%d/%s", instance.Engine, instance.Username, instance.Password, *dbAddr, *dbPort, instance.DBName)
		errChan <- nil
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