package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/mitchellh/goamz/aws"
    "io/ioutil"
    "os"
)

var configFile string

type syncConfig struct {
    RsOauthToken string `json:"rs_oauth_token"`
    RsEndPoint   string `json:"rs_endpoint,omitempty"`
    AwsAccessKey string `json:"aws_access_key,omitempty"`
    AwsSecretKey string `json:"aws_secret_key,omitepty"`
    AwsRegion    string `json:"aws_region,omitempty"`
}

func loadAwsFromProvidedConfig(loadedConfig *syncConfig) (auth aws.Auth, err error) {
    fmt.Printf("Attempting to load config from '%v'\n",configFile)
    if loadedConfig.AwsAccessKey == "" && loadedConfig.AwsSecretKey == "" {
        err = errors.New("No AWS credentials were found in" + configFile)
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
    configFile = configFilePath
    file, e := ioutil.ReadFile(configFilePath)
    if e != nil {
        fmt.Printf("Could not: %v\n", e)
        os.Exit(1)
    }
    var syncConfig syncConfig
    json.Unmarshal(file, &syncConfig)
    return syncConfig
}