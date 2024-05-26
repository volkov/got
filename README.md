# GOT
Teamcity console client on go

## Usage
```bash
# set environment variables for teamcity
export TEAMCITY_URL="http://teamcity:8111"
export TEAMCITY_LOGIN="user"
export TEAMCITY_PASSWORD="password"
```

Help
```bash
go run main.go --help
```

Run build and wait for finish
```bash
go run main.go --command=build --id=BackendBuild
```
