package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v1"
)

//http://stackoverflow.com/questions/19253469/make-a-url-encoded-post-request-using-http-newrequest
var p = fmt.Println

func syncResources(rsResources []rsResource, awsResources []awsResource, resourceType string) {
	for _, aws_resource := range awsResources {
		for _, rs_resource := range rsResources {
			if rs_resource.ResourceID == aws_resource.ID {
				href := rsExtractRsResourceHref(rs_resource.Links)
				queryParam := rsGetUpdateParam(resourceType)
				rsName := rs_resource.Name
				awsName := aws_resource.Name
				if rsName == awsName {
					fmt.Printf("No need to update Rightscale %v from %v -to- %v\n", resourceType, rsName, awsName)
					continue
				}
				updateParams := rsUpdateParams{href: href, queryParam: queryParam, oldValue: rsName, newValue: awsName}
				returnValue := rsUpdate(updateParams)
				if returnValue {
					fmt.Printf("Rightscale %v was updated from %v -to- %v\n", resourceType, rsName, awsName)
				} else {
					fmt.Printf("Error: coult not update Rightscale %v from %v -to- %v\n", resourceType, rsName, awsName)
				}
			}
		}
	}
}

func main() {
	vpcID := kingpin.Flag("vpc", "Id of the VPC to sync with Rightscale").Required().Short('v').String()
	configFile := kingpin.Flag("config", "JSON configuration file with the following keys\n rs_oauth_token - (required)\n rs_endpoint - (optional, defaults to my.rightscale.com)\n aws_region - (required)\n aws_access_key - (optional, ENV or aws config file will used)\n aws_secret_key - (optional, ENV or aws config file will used)").Short('f').Default("config.json").String()
	kingpin.Parse()
	loadedConfig := loadConfigFile(*configFile)
	loadCredentials(&loadedConfig)
	p("Preparing resources for sync...")
	rsNetworks := rsNetworks()
	rs_network := rsSelectNetwork(rsNetworks, *vpcID)
	aws_network := awsVpc(*vpcID)
	p("Beginning sync\n")
	syncResources([]rsResource{rs_network}, []awsResource{aws_network}, "network")
	vpcRsHref := rsExtractRsResourceHref(rs_network.Links)
	rs_route_tables := rsRouteTables(vpcRsHref)
	aws_route_tables := awsRouteTable(*vpcID)
	syncResources(rs_route_tables, aws_route_tables, "routetable")
	rs_internet_gateways := rsInternetGateways(vpcRsHref)
	aws_internet_gateways := awsIgw(*vpcID)
	syncResources(rs_internet_gateways, []awsResource{aws_internet_gateways}, "gateway")
	rs_subnets := rsSubnets(vpcRsHref)
	aws_subnets := awsSubnets(*vpcID)
	syncResources(rs_subnets, aws_subnets, "subnet")
}
