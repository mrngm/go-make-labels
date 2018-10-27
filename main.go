// Set `OCTOKIT_ACCESS_TOKEN` environment variable and provide an argument to
// point to your target repository (ex "rails/rails", "kylemacey/go-make-labels")

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/octokit/go-octokit/octokit"
)

func main() {
	jsonFile := flag.String("json", "example.json", "Use this JSON file for user labels")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s [-json file] <user/repository>\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Don't forget to set the environment variable OCTOKIT_ACCESS_TOKEN.")
		return
	}

	client := octokit.NewClient(getAuthMethod())

	labels, result := client.Labels().All(nil, getRepoParams())

	handleRequestError(result)

	userlabels := make([]UserLabel, 0)

	json.Unmarshal(getJsonFromFile(*jsonFile), &userlabels)

	for i := 0; i < len(userlabels); i++ {
		if contains(labels, userlabels[i].Name) {
			updateLabel(userlabels[i], client)
		} else {
			createLabel(userlabels[i], client)
		}
	}
}

func getAuthMethod() octokit.TokenAuth {
	return octokit.TokenAuth{AccessToken: os.Getenv("OCTOKIT_ACCESS_TOKEN")}
}

func handleRequestError(result *octokit.Result) {
	res := *result
	if res.HasError() {
		fmt.Println(res.Error())
	}
}

func getRepoParams() octokit.M {
	owner, repo := getOwnerAndRepoName()
	return octokit.M{"owner": owner, "repo": repo}
}

func getLabelParams(labelName string) octokit.M {
	owner, repo := getOwnerAndRepoName()
	return octokit.M{"owner": owner, "repo": repo, "name": labelName}
}

func getOwnerAndRepoName() (string, string) {
	owner_repo_name := strings.Split(os.Args[1], "/")
	return owner_repo_name[0], owner_repo_name[1]
}

func updateLabel(userLabel UserLabel, octokitClient *octokit.Client) {
	client := *octokitClient
	fmt.Println("Updating `" + userLabel.Name + "`...")
	_, result := client.Labels().Update(nil, getLabelParams(userLabel.Name), octokit.M{
		"color": userLabel.Color,
		// There is no support in go-octokit yet for description, but it doesn't
		// hurt to leave it here until it gets implemented
		"description": userLabel.Description,
	})
	handleRequestError(result)
}

func createLabel(userLabel UserLabel, octokitClient *octokit.Client) {
	client := *octokitClient
	fmt.Println("Creating `" + userLabel.Name + "`...")
	_, result := client.Labels().Create(nil, getRepoParams(), octokit.M{
		"name":  userLabel.Name,
		"color": userLabel.Color,
		// There is no support in go-octokit yet for description, but it doesn't
		// hurt to leave it here until it gets implemented
		"description": userLabel.Description,
	})
	handleRequestError(result)
}

func contains(arr []octokit.Label, str string) bool {
	for i := 0; i < len(arr); i++ {
		if strings.ToLower(arr[i].Name) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func getJsonFromFile(path string) []byte {
	jsonFile, err := os.Open(path)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	return byteValue
}

type UserLabel struct {
	Name        string
	Color       string
	Description string
}
