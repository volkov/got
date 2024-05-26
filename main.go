package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Project struct {
	Name string `json:"name"`
}

type Projects struct {
	Count    int       `json:"count"`
	Projects []Project `json:"project"`
}

func main() {
	command := flag.String("command", "list", "Command to execute")
	flag.Parse()
	switch *command {
	case "list":
		listProjects()
	default:
		fmt.Println("Unknown command")
	}
}

func listProjects() {
	host := os.Getenv("TEAMCITY_HOST")
	login := os.Getenv("TEAMCITY_LOGIN")
	password := os.Getenv("TEAMCITY_PASSWORD")

	url := fmt.Sprintf("https://%s/httpAuth/app/rest/projects", host)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(login, password)
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
	// print each project on a new line
	for _, project := range projects.Projects {
		fmt.Println(project.Name)
	}
}
