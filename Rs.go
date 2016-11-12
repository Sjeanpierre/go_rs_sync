package main

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
)

var bearerToken string
var rsEndPoint string
var rsRefreshToken string

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

func rsNetworks() []rsResource {
    url := "/api/networks.json"
    RequestParams := rsRequestParams{method: "GET", url: url}
    RSResponse := rsRequest(RequestParams)
    var networks []rsResource
    NetworkJSON := []byte(RSResponse)
    json.Unmarshal(NetworkJSON, &networks)
    return networks
}

func rsSelectNetwork(rsNetworkList []rsResource, vpcID string) (rsNetwork rsResource) {
    for _, network := range rsNetworkList {
        if network.ResourceID == vpcID {
            rsNetwork = network
            break
        }
    }
    if rsNetwork.ResourceID != vpcID {
        p("Could not locate specified VPC in Rightscale")
        os.Exit(1)
    }
    return
}

func rsExtractRsResourceHref(resourceLinks []rsLinks) (resourceHref string) {
    for _, link := range resourceLinks {
        if link.Rel == "self" {
            resourceHref = link.Href
            break
        }
    }
    return
}

func rsExtractNetworkCloudHref(resourceLinks []rsLinks) (resourceHref string) {
    for _, link := range resourceLinks {
        if link.Rel == "cloud" {
            resourceHref = link.Href
            break
        }
    }
    return
}

func rsSubnets(networkHref string, cloudHref string) (subnets []rsResource) {
    cloudHrefParts := strings.Split(cloudHref,"/")
    cloudID := cloudHrefParts[len(cloudHrefParts)-1]
    url := "/api/clouds/" +cloudID+ "/subnets.json"
    filter := "?filter[]=network_href=="
    fullURL := strings.Join([]string{url, filter, networkHref}, "")
    RequestParams := rsRequestParams{method: "GET", url: fullURL}
    RSResponse := rsRequest(RequestParams)
    SubnetJSON := []byte(RSResponse)
    json.Unmarshal(SubnetJSON, &subnets)
    return
}

func rsInternetGateways(networkHref string) (internetGateways []rsResource) {
    url := "/api/network_gateways.json"
    filter := "?filter[]=network_href=="
    fullURL := strings.Join([]string{url, filter, networkHref}, "")
    RequestParams := rsRequestParams{method: "GET", url: fullURL}
    RSResponse := rsRequest(RequestParams)
    InternetGatewayJSON := []byte(RSResponse)
    json.Unmarshal(InternetGatewayJSON, &internetGateways)
    return
}

func rsRouteTables(networkHref string) (routeTables []rsResource) {
    url := "/api/route_tables.json"
    filter := "?filter[]=network_href=="
    fullURL := strings.Join([]string{url, filter, networkHref}, "")
    RequestParams := rsRequestParams{method: "GET", url: fullURL}
    RSResponse := rsRequest(RequestParams)
    RouteTableJSON := []byte(RSResponse)
    json.Unmarshal(RouteTableJSON, &routeTables)
    return
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

func rsGetUpdateParam(resourceType string) (updateParam string) {
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
    return
}
