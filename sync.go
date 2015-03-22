package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//http://stackoverflow.com/questions/19253469/make-a-url-encoded-post-request-using-http-newrequest
var p = fmt.Println
var bearerToken string
var rsEndPoint string
var rsRefreshToken string
var awsCredentials aws.Auth
var awsRegion aws.Region

type rsOauth2 struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type rsLinks struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type rsResource struct {
	ResourceID string    `json:"resource_uid"`
	Name       string    `json:"name"`
	Links      []rsLinks `json:"links"`
}

type rsRequestParams struct {
	method string
	url    string
}

type rsUpdateParams struct {
	href       string
	queryParam string
	newValue   string
	oldValue   string
}

type awsResource struct {
	ID   string
	Name string
}

type syncConfig struct {
	RsOauthToken string `json:"rs_oauth_token"`
	RsEndPoint   string `json:"rs_endpoint,omitempty"`
	AwsAccessKey string `json:"aws_access_key,omitempty"`
	AwsSecretKey string `json:"aws_secret_key,omitepty"`
	AwsRegion    string `json:"aws_region,omitempty"`
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

func rsNetworks() []rsResource {
	url := "/api/networks.json"
	RequestParams := rsRequestParams{method: "GET", url: url}
	RSResponse := rsRequest(RequestParams)
	var networks []rsResource
	NetworkJSON := []byte(RSResponse)
	json.Unmarshal(NetworkJSON, &networks)
	return networks
}

func rsSelectNetwork(rsNetworkList []rsResource, vpcID string) rsResource {
	var rsNetwork rsResource
	for _, network := range rsNetworkList {
		if network.ResourceID == vpcID {
			rsNetwork = network
			break
		}
	}
	return rsNetwork
}

func rsExtractRsResourceHref(resourceLinks []rsLinks) string {
	resourceHref := ""
	for _, link := range resourceLinks {
		if link.Rel == "self" {
			resourceHref = link.Href
			break
		}
	}
	return resourceHref
}

func rsSubnets(networkHref string) []rsResource {
	url := "/api/clouds/1/subnets.json"
	filter := "?filter[]=network_href=="
	fullURL := strings.Join([]string{url, filter, networkHref}, "")
	RequestParams := rsRequestParams{method: "GET", url: fullURL}
	RSResponse := rsRequest(RequestParams)
	var subnets []rsResource
	SubnetJSON := []byte(RSResponse)
	json.Unmarshal(SubnetJSON, &subnets)
	return subnets
}

func rsInternetGateways(networkHref string) []rsResource {
	url := "/api/network_gateways.json"
	filter := "?filter[]=network_href=="
	fullURL := strings.Join([]string{url, filter, networkHref}, "")
	RequestParams := rsRequestParams{method: "GET", url: fullURL}
	RSResponse := rsRequest(RequestParams)
	var internetGateways []rsResource
	InternetGatewayJSON := []byte(RSResponse)
	json.Unmarshal(InternetGatewayJSON, &internetGateways)
	return internetGateways
}

func rsRouteTables(networkHref string) []rsResource {
	url := "/api/route_tables.json"
	filter := "?filter[]=network_href=="
	fullURL := strings.Join([]string{url, filter, networkHref}, "")
	RequestParams := rsRequestParams{method: "GET", url: fullURL}
	RSResponse := rsRequest(RequestParams)
	var routetables []rsResource
	RouteTableJSON := []byte(RSResponse)
	json.Unmarshal(RouteTableJSON, &routetables)
	return routetables
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

func rsBearerToken() string {
	refreshToken := rsRefreshToken
	data := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}
	client := http.Client{}
	url := strings.Join([]string{rsEndPoint, "/api/oauth2"}, "")
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))

	if err != nil {
		p("we ran into an error 1")
		p(err)
		os.Exit(1)
	}
	req.Header.Add("X_API_VERSION", "1.5")
	req.Header.Add("accept", "json")
	response, error := client.Do(req)

	if error != nil {
		p("we ran into an error")
		p(error)
		os.Exit(1)
	}
	defer response.Body.Close()
	ResponseText, error := ioutil.ReadAll(response.Body)
	if error != nil {
		p("there has been an issue")
		os.Exit(1)
	}
	result := rsOauth2{}
	json.Unmarshal([]byte(ResponseText), &result)
	token := strings.Join([]string{"Bearer", result.AccessToken}, " ")
	return token
}

func rsRequest(RequestParams rsRequestParams) []byte {

	if bearerToken == "" {
		bearerToken = rsBearerToken()
	}
	client := http.Client{}
	url := strings.Join([]string{rsEndPoint, RequestParams.url}, "")
	req, err := http.NewRequest(RequestParams.method, url, nil)

	if err != nil {
		p("we ran into an error 1")
		p(err)
		os.Exit(1)
	}
	req.Header.Add("X_API_VERSION", "1.5")
	req.Header.Add("Authorization", bearerToken)
	req.Header.Add("accept", "json")
	response, error := client.Do(req)

	if error != nil {
		p("we ran into an error")
		p(error)
		os.Exit(1)
	}
	defer response.Body.Close()
	ReponseText, error := ioutil.ReadAll(response.Body)
	if error != nil {
		p("there has been an issue")
		os.Exit(1)
	}
	return ReponseText
}

func rsUpdate(UpdateParams rsUpdateParams) bool {

	if bearerToken == "" {
		bearerToken = rsBearerToken()
	}
	client := http.Client{}
	url := strings.Join([]string{rsEndPoint, UpdateParams.href, UpdateParams.queryParam, UpdateParams.newValue}, "")
	req, err := http.NewRequest("PUT", url, nil)

	if err != nil {
		p("we ran into an error 1")
		p(err)
		os.Exit(1)
	}
	req.Header.Add("X_API_VERSION", "1.5")
	req.Header.Add("Authorization", bearerToken)
	req.Header.Add("accept", "json")
	response, error := client.Do(req)

	if error != nil {
		p("we ran into an error")
		p(error)
		os.Exit(1)
	}
	defer response.Body.Close()
	if response.StatusCode == 204 {
		return true
	}
	return false
}

func rsGetUpdateParam(resourceType string) string {
	updateParam := ""
	switch resourceType {
	case "network":
		updateParam = "?network[name]="
	case "subnet":
		updateParam = "?subnet[name]="
	case "gateway":
		updateParam = "?network_gateway[name]="
	case "routetable":
		updateParam = "?route_table[name]="
	}
	return updateParam
}

func syncResources(rsResources []rsResource, awsResources []awsResource, resourceType string) {
	for _, aws_resource := range awsResources {
		for _, rs_resource := range rsResources {
			if rs_resource.ResourceID == aws_resource.ID {
				href := rsExtractRsResourceHref(rs_resource.Links)
				queryParam := rsGetUpdateParam(resourceType)
				rsName := rs_resource.Name
				awsName := aws_resource.Name
				updateParams := rsUpdateParams{href: href, queryParam: queryParam, oldValue: rsName, newValue: awsName}
				returnValue := rsUpdate(updateParams)
				if returnValue {
					fmt.Printf("Rightscale %v was updated from %v to %v\n", resourceType, rsName, awsName)
				} else {
					fmt.Printf("Error: coult not update Rightscale %v from %v to %v\n", resourceType, rsName, awsName)
				}
			}
		}
	}
}

func loadAwsFromProvidedConfig(loadedConfig *syncConfig) (auth aws.Auth, err error) {
	p("Attempting to load aws config from provided config file\n")
	if loadedConfig.AwsAccessKey == "" && loadedConfig.AwsSecretKey == "" {
		err = errors.New("No AWS credentials were found in provided config file\n")
	}
	auth = aws.Auth{AccessKey: loadedConfig.AwsAccessKey, SecretKey: loadedConfig.AwsSecretKey}
	return
}

func loadAwsDefaultConfig() (auth aws.Auth, err error) {
	p("Attempting to load AWS auth from ~/.aws/credentials file")
	auth, err = aws.SharedAuth()
	return
}

func loadAwsEnvironmentCreds() (auth aws.Auth, err error) {
	p("Attempting to load aws credentials from environment varibles")
	auth, err = aws.EnvAuth()
	return
}

func loadAwsUserProvidedCreds() (auth aws.Auth, err error) {
	p("Please supply AWS credentials")
	p("AWS Access Key: ")
	var accessKey string
	fmt.Scanln(&accessKey)
	p("AWS Secret Key: ")
	var secretKey string
	fmt.Scanln(&secretKey)
	auth = aws.Auth{AccessKey: accessKey, SecretKey: secretKey}
	return
}

func loadCredentials(loadedConfig *syncConfig) {
	if loadedConfig.RsOauthToken != "" {
		rsRefreshToken = loadedConfig.RsOauthToken
	} else {
		p("Rightscale Oauth token not provided!")
		os.Exit(1)
	}
	if loadedConfig.RsEndPoint != "" {
		rsEndPoint = loadedConfig.RsEndPoint
	} else {
		rsEndPoint = "https://my.rightscale.com"
	}
	if loadedConfig.AwsRegion != "" {
		awsRegion = aws.Regions[loadedConfig.AwsRegion]
	} else {
		p("AWS region was not provided")
		os.Exit(1)
	}
	auth, err := loadAwsFromProvidedConfig(loadedConfig)
	if err == nil {
		awsCredentials = auth
		return
	}
	p(err)
	auth, err = loadAwsDefaultConfig()
	if err == nil {
		awsCredentials = auth
		return
	}
	p(err)
	auth, err = loadAwsEnvironmentCreds()
	if err == nil {
		awsCredentials = auth
		return
	}
	auth, err = loadAwsUserProvidedCreds()
	if err == nil {
		awsCredentials = auth
		return
	}
	p(err)
	os.Exit(1)
}

func loadConfigFile(configFilePath string) syncConfig {
	file, e := ioutil.ReadFile(configFilePath)
	if e != nil {
		fmt.Printf("Could not: %v\n", e)
		os.Exit(1)
	}
	var syncConfig syncConfig
	json.Unmarshal(file, &syncConfig)
	return syncConfig
}

func main() {
	vpcID := kingpin.Flag("vpc", "Id of the VPC to sync with Rightscale").Required().Short('v').String()
	configFile := kingpin.Flag("config", "JSON configuration file with the following keys\n rs_oauth_token - (required)\n rs_endpoint - (optional, defaults to my.rightscale.com)\n aws_region - (required)\n aws_access_key - (optional, ENV or aws config file will used)\n aws_secret_key - (optional, ENV or aws config file will used)").Short('f').Default("config.json").String()
	kingpin.Parse()
	loadedConfig := loadConfigFile(*configFile)
	loadCredentials(&loadedConfig)
	p(vpcID)
	rsNetworks := rsNetworks()
	rs_network := rsSelectNetwork(rsNetworks, *vpcID)
	aws_network := awsVpc(*vpcID)
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
