# go_rs_sync
Go utility written to synchronize AWS vpc meta-data with Rightscale

My first go project, so if you come across any inefficiencies please let me know.

Supports the syncing of 
* VPC Names
* Subnet Names
* Route Names
* Internet Gateway Names

```
usage: go_rs_sync --vpc=VPC [<flags>]
Flags:
  --help         Show help.
  -v, --vpc=VPC  Id of the VPC to sync with Rightscale
  -f, --config="config.json" JSON configuration file with the following keys
      rs_oauth_token - (required)
      rs_endpoint - (optional, defaults to my.rightscale.com)
      aws_region - (required)
      aws_access_key - (optional,if not provided ENV varibles or ~/.aws/credentials will be used)
      aws_secret_key - (optional,if not provided ENV varibles or ~/.aws/credentials will be used)
```
