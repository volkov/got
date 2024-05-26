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
	"time"
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
	case "wait":
		waitForBuildToFinish(c, *id)
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
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if err == nil {
			err = closeErr
		}
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if err == nil {
			err = closeErr
		}
	}(resp.Body)

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
	buildId, done := startBuild(c, id, branch)
	if done {
		return
	}
	fmt.Println("Build started ", buildId, "...")
	//wait
	waitForBuildToFinish(c, fmt.Sprintf("%d", buildId))

}

func startBuild(c Credentials, id string, branch string) (int, bool) {
	url := fmt.Sprintf("https://%s/httpAuth/app/rest/buildQueue", c.Host)

	req, err := http.NewRequest("POST", url, strings.NewReader(formRequestBody(branch, id)))
	if err != nil {
		fmt.Println(err)
		return 0, true
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, true
	}

	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if err == nil {
			err = closeErr
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0, true
	}
	//print resp body
	fmt.Println(string(respBody))

	// Get the build ID from the response
	var build struct {
		Id int `json:"id"`
	}
	err = json.Unmarshal(respBody, &build)
	if err != nil {
		fmt.Println(err)
		return 0, true
	}

	buildId := build.Id
	return buildId, false
}

func formRequestBody(branch string, id string) string {
	// Form the XML body
	var requestBody string
	if branch != "" {
		requestBody = fmt.Sprintf(`<build branchName="%s"><buildType id="%s"/></build>`, branch, id)
	} else {
		requestBody = fmt.Sprintf(`<build><buildType id="%s"/></build>`, id)
	}
	return requestBody
}

func waitForBuildToFinish(c Credentials, buildId string) {
	buildURL := fmt.Sprintf("https://%s/httpAuth/app/rest/builds/id:%s", c.Host, buildId)

	// Poll the build status until the build is finished
	for {
		state, done := getState(c, buildURL)
		if done {
			return
		}
		fmt.Println("Build state:", state)
		if state != "running" && state != "queued" {
			break
		}

		// Wait for a while before polling again
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Build finished")
}

func getState(c Credentials, buildURL string) (string, bool) {
	req, err := http.NewRequest("GET", buildURL, nil)
	if err != nil {
		fmt.Println(err)
		return "", true
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", true
	}

	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if err == nil {
			err = closeErr
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", true
	}

	var build struct {
		State string `json:"state"`
	}
	err = json.Unmarshal(body, &build)
	if err != nil {
		fmt.Println(err)
		return "", true
	}
	state := build.State
	return state, false
}
