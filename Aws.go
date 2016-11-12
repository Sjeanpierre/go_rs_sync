package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

//var awsCredentials aws.Auth
var awsRegion aws.Config //set with aws.NewConfig().WithRegion("us-west-2") in config.go

type awsResource struct {
	ID   string
	Name string
}

func awsExtractName(tags []*ec2.Tag) (name string) {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			name = *tag.Value
			break
		}
	}
	return
}

func awsClient() (client *ec2.EC2) {
	client = ec2.New(session.New(), &awsRegion) //todo, error handling here
	return
}

func awsSubnets(vpcid string) []awsResource {

	client := awsClient()
	params := ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{&vpcid},
			},
		},
	}

	response, err := client.DescribeSubnets(&params)
	if err != nil {
		log.Fatalf("Issue encountered retrieving subnets, %s",err)
	}
	subnets := &response.Subnets
	var ReturnedSubs []awsResource
	for _, subnet := range *subnets {
		subnetID := *subnet.SubnetId
		subnetName := awsExtractName(subnet.Tags)
		subnetDetails := awsResource{ID: subnetID, Name: subnetName}
		ReturnedSubs = append(ReturnedSubs, subnetDetails)
	}
	return ReturnedSubs
}

func awsVpc(vpcid string) awsResource {
	client := awsClient()
	vpcids := []*string{&vpcid}
	response, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{VpcIds: vpcids})
	if err != nil {
		log.Fatalf("Ecountered issue retrieving VPCs, %s",err) // todo, better error handling
	}
	vpc := response.Vpcs[0]
	vpcName := awsExtractName(vpc.Tags)
	vpcDetails := awsResource{ID: *vpc.VpcId, Name: vpcName}
	return vpcDetails
}

func awsIgw(vpcid string) (igwDetails awsResource) {
	client := awsClient()
	params := ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.vpc-id"),
				Values: []*string{&vpcid},
			},
		},
	}
	response, err := client.DescribeInternetGateways(&params)
	if err != nil {
		log.Fatalf("Ecountered issue retrieving Internet Gateways, %s", err)
	}
	igw := response.InternetGateways[0]
	igwName := awsExtractName(igw.Tags)
	igwDetails = awsResource{ID: *igw.InternetGatewayId, Name: igwName}
	return
}

// todo, extract filtering out to another function
func awsRouteTable(vpcid string) (ReturnedRouteTables []awsResource) {
	client := awsClient()
	params := ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{&vpcid},
			},
		},
	}
	response, err := client.DescribeRouteTables(&params)
	if err != nil {
		log.Fatalf("Ecountered issue retrieving Route Table details, %s",err)
	}
	routeTables := response.RouteTables
	for _, routeTable := range routeTables {
		routeTableID := routeTable.RouteTableId
		routeTableName := awsExtractName(routeTable.Tags)
		routeTableDetails := awsResource{ID: *routeTableID, Name: routeTableName}
		ReturnedRouteTables = append(ReturnedRouteTables, routeTableDetails)
	}
	return
}