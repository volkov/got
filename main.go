package main

import "C"
import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Projects struct {
	Count    int       `json:"count"`
	Projects []Project `json:"project"`
}

func main() {
	command := flag.String("command", "list", "Command to execute")
	id := flag.String("id", "", "ID of the project or configuration to use")
	branch := flag.String("branch", "", "Branch to use")
	flag.Parse()

	c := NewCredentials()

	switch *command {
	case "list":
		listProjects(c)
	case "list-builds":
		listConfigurations(c, *id)
	case "build":
		buildConfiguration(c, *id, *branch)
	default:
		fmt.Println("Unknown command")
	}
}

type Credentials struct {
	Username string
	Password string
	Host     string
}

func NewCredentials() Credentials {
	return Credentials{
		Host:     os.Getenv("TEAMCITY_HOST"),
		Username: os.Getenv("TEAMCITY_LOGIN"),
		Password: os.Getenv("TEAMCITY_PASSWORD"),
	}
}

type Project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func listProjects(c Credentials) {
	url := fmt.Sprintf("https://%s/httpAuth/app/rest/projects", c.Host)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var projects Projects
	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println(err)
		return
	}
	// print each project with id on a new line
	for _, project := range projects.Projects {
		fmt.Println("ID:", project.Id, " Name:", project.Name)
	}
}

type BuildType struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ProjectName string `json:"projectName"`
	ProjectId   string `json:"projectId"`
	Href        string `json:"href"`
	WebUrl      string `json:"webUrl"`
}

type Configurations struct {
	Count      int         `json:"count"`
	BuildTypes []BuildType `json:"buildType"`
}

func listConfigurations(c Credentials, id string) {
	url := fmt.Sprintf("https://%s/httpAuth/app/rest/projects/id:%s/buildTypes", c.Host, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var configurations Configurations
	err = json.Unmarshal(body, &configurations)
	if err != nil {
		fmt.Println(err)
		return
	}

	// print each configuration with id on a new line
	for _, configuration := range configurations.BuildTypes {
		fmt.Println("ID:", configuration.Id, " Name:", configuration.Name)
	}
}

func buildConfiguration(c Credentials, id, branch string) {
	url := fmt.Sprintf("https://%s/httpAuth/app/rest/buildQueue", c.Host)

	// Form the XML body
	var body string
	if branch != "" {
		body = fmt.Sprintf(`<build branchName="%s"><buildType id="%s"/></build>`, branch, id)
	} else {
		body = fmt.Sprintf(`<build><buildType id="%s"/></build>`, id)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/xml")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Build started")
}
