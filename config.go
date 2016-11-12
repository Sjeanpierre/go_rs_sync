package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "github.com/aws/aws-sdk-go/aws"
    "log"
)


type syncConfig struct {
    RsOauthToken string `json:"rs_oauth_token"`
    RsEndPoint   string `json:"rs_endpoint,omitempty"`
    AwsRegion    string `json:"aws_region,omitempty"`
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
        awsRegion = *aws.NewConfig().WithRegion(loadedConfig.AwsRegion) //aws.Config{Region: aws.String(loadedConfig.AwsRegion)}
        log.Printf("Using AWS region %s",*awsRegion.Region)
    } else {
        p("AWS region was not provided")
        os.Exit(1)
    }
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