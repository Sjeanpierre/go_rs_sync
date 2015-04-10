package main

import (
    "github.com/mitchellh/goamz/ec2"
    "github.com/mitchellh/goamz/aws"
    "os"
)

var awsCredentials aws.Auth
var awsRegion aws.Region

type awsResource struct {
    ID   string
    Name string
}

func awsExtractName(tags []ec2.Tag) string {
    name := ""
    for _, tag := range tags {
        if tag.Key == "Name" {
            name = tag.Value
            break
        }
    }
    return name
}

func awsClient() *ec2.EC2 {
    client := ec2.New(awsCredentials, awsRegion)

    return client
}

func awsSubnets(vpcid string) []awsResource {

    client := awsClient()

    filter := ec2.NewFilter()
    filter.Add("tag-key", "Name")
    filter.Add("vpc-id", vpcid)

    response, err := client.DescribeSubnets(nil, filter)
    if err != nil {
        p("ran into issue trying to get subnets")
        p(err)
    }
    subnets := response.Subnets
    var ReturnedSubs []awsResource
    for _, subnet := range subnets {
        subnetID := subnet.SubnetId
        subnetName := awsExtractName(subnet.Tags)
        subnetDetails := awsResource{ID: subnetID, Name: subnetName}
        ReturnedSubs = append(ReturnedSubs, subnetDetails)
    }
    return ReturnedSubs
}

func awsVpc(vpcid string) awsResource {
    client := awsClient()
    vpcids := []string{vpcid}
    response, err := client.DescribeVpcs(vpcids, nil)
    if err != nil {
        p("ran into issue trying to get vpc details")
        p(err)
        os.Exit(1)
    }
    vpc := response.VPCs[0]
    vpcName := awsExtractName(vpc.Tags)
    vpcDetails := awsResource{ID: vpc.VpcId, Name: vpcName}
    return vpcDetails
}

func awsIgw(vpcid string) awsResource {
    client := awsClient()
    filter := ec2.NewFilter()
    filter.Add("attachment.vpc-id", vpcid)
    response, err := client.DescribeInternetGateways(nil, filter)
    if err != nil {
        p("ran into issue trying to get IGW details")
        p(err)
    }
    igw := response.InternetGateways[0]
    igwName := awsExtractName(igw.Tags)
    igwDetails := awsResource{ID: igw.InternetGatewayId, Name: igwName}
    return igwDetails
}

func awsRouteTable(vpcid string) []awsResource {
    client := awsClient()
    filter := ec2.NewFilter()
    filter.Add("vpc-id", vpcid)
    filter.Add("tag-key", "Name")
    response, err := client.DescribeRouteTables(nil, filter)
    if err != nil {
        p("ran into issue trying to get Route Table details")
        p(err)
    }
    routeTables := response.RouteTables
    var ReturnedRouteTables []awsResource
    for _, routeTable := range routeTables {
        routeTableID := routeTable.RouteTableId
        routeTableName := awsExtractName(routeTable.Tags)
        routeTableDetails := awsResource{ID: routeTableID, Name: routeTableName}
        ReturnedRouteTables = append(ReturnedRouteTables, routeTableDetails)
    }
    return ReturnedRouteTables
}