package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
)

//http://stackoverflow.com/questions/19253469/make-a-url-encoded-post-request-using-http-newrequest
var p = log.Println

func syncResources(rsResources []rsResource, awsResources []awsResource, resourceType string) {
	for _, awsResource := range awsResources {
		for _, rsResource := range rsResources {
			if rsResource.ResourceID == awsResource.ID {
				href := rsExtractRsResourceHref(rsResource.Links)
				queryParam := rsGetUpdateParam(resourceType)
				rsName := rsResource.Name
				awsName := awsResource.Name
				if rsName == awsName {
					log.Printf("No need to update Rightscale %s from: %s to: %s\n", resourceType, rsName, awsName)
					continue
				}
				updateParams := rsUpdateParams{href: href, queryParam: queryParam, oldValue: rsName, newValue: awsName}
				returnValue := rsUpdate(updateParams)
				if returnValue {
					log.Printf("Rightscale %s was updated from: %s to: %s\n", resourceType, rsName, awsName)
				} else {
					log.Printf("Error: coult not update Rightscale %s from: %s to: %s\n", resourceType, rsName, awsName)
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
	p("Using VPC:", *vpcID)
	rsNetwork := rsSelectNetwork(rsNetworks, *vpcID)
	awsNetwork := awsVpc(*vpcID)
	p("Beginning sync")
	p("Syncing top level network")
	syncResources([]rsResource{rsNetwork}, []awsResource{awsNetwork}, "network")
	vpcRsHref := rsExtractRsResourceHref(rsNetwork.Links)
	rsRouteTables := rsRouteTables(vpcRsHref)
	awsRouteTables := awsRouteTable(*vpcID)
	p("Syncing Route Tables")
	syncResources(rsRouteTables, awsRouteTables, "routetable")
	rsInternetGateways := rsInternetGateways(vpcRsHref)
	awsInternetGateways := awsIgw(*vpcID)
	p("Syncing Internet Gateways")
	syncResources(rsInternetGateways, []awsResource{awsInternetGateways}, "gateway")
	vpcCloudHref := rsExtractNetworkCloudHref(rsNetwork.Links)
	rsSubnets := rsSubnets(vpcRsHref, vpcCloudHref)
	awsSubnets := awsSubnets(*vpcID)
	p("Syncing Subnets")
	syncResources(rsSubnets, awsSubnets, "subnet")
	log.Print("done")

}
