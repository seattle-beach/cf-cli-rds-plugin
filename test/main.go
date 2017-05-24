package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/rds"
//	"encoding/json"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config {
		Region: aws.String(endpoints.UsEast1RegionID),
	}))

	svc := rds.New(sess)

	subnetGroupParams := &rds.DescribeDBSubnetGroupsInput{
		Filters: []*rds.Filter{
			{ // Required
				Name: aws.String("*"), // Required
				Values: []*string{ // Required
					aws.String("*"), // Required
					// More values...
				},
			},
			// More values...
		},
		Marker:     aws.String("String"),
		MaxRecords: aws.Int64(20),
	}
	subnetGroupsResp, err := svc.DescribeDBSubnetGroups(subnetGroupParams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
//	fmt.Println(resp.String())
	subnetGroupName := *subnetGroupsResp.DBSubnetGroups[0].DBSubnetGroupName

	params := &rds.CreateDBInstanceInput{
		DBInstanceClass:         aws.String("db.t2.micro"), // Required
		DBInstanceIdentifier:    aws.String("test-db-instance-2"), // Required
		Engine:                  aws.String("postgres"), // Required
		AllocatedStorage:        aws.Int64(5),
		AutoMinorVersionUpgrade: aws.Bool(true),
		AvailabilityZone:        aws.String("us-east-1a"),
		BackupRetentionPeriod:   aws.Int64(1),
		//CharacterSetName:        aws.String("String"),
		CopyTagsToSnapshot:      aws.Bool(true),
		//DBClusterIdentifier:     aws.String("String"),
		DBName:                  aws.String("testdb"),
		DBParameterGroupName:    aws.String("default.postgres9.6"),
		//DBSecurityGroups: []*string{
		//	aws.String("default"), // Required
		//},
		DBSubnetGroupName:               aws.String(subnetGroupName),
		//Domain:                          aws.String("String"),
		//DomainIAMRoleName:               aws.String("String"),
		//EnableIAMDatabaseAuthentication: aws.Bool(true),
		//EngineVersion:                   aws.String("String"),
		//Iops:                            aws.Int64(1),
		//KmsKeyId:                        aws.String("String"),
		//LicenseModel:                    aws.String("String"),
		MasterUserPassword:              aws.String("password"),
		MasterUsername:                  aws.String("root"),
		//MonitoringInterval:              aws.Int64(1),
		//MonitoringRoleArn:               aws.String("String"),
		MultiAZ:                         aws.Bool(false),
		//OptionGroupName:                 aws.String("String"),
		Port:                            aws.Int64(5432),
		//PreferredBackupWindow:      aws.String("String"),
		//PreferredMaintenanceWindow: aws.String("String"),
		//PromotionTier:              aws.Int64(1),
		PubliclyAccessible:         aws.Bool(true),
		//StorageEncrypted:           aws.Bool(true),
		//StorageType:                aws.String("String"),
		Tags: []*rds.Tag{
			{ // Required
				Key:   aws.String("test-db"),
				Value: aws.String("test-db-guid"),
			},
			// More values...
		},
		//TdeCredentialArn:      aws.String("String"),
		//TdeCredentialPassword: aws.String("String"),
		//Timezone:              aws.String("String"),
		//VpcSecurityGroupIds: []*string{
		//	aws.String("sg-a8d2a9d6"), // Required
		//	// More values...
		//},
	}
	dbResp, err := svc.CreateDBInstance(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(dbResp)
}