package jira

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mkobaly/devop/config"
)

//ApiError represents a set of error(s) that the rest api threw
type ApiError struct {
	ErrorMessages []string    `json:"errorMessages"`
	Errors        interface{} `json:"errors"`
}

func (e ApiError) String() string {
	return strings.Join(e.ErrorMessages, " ")
}

type BaseFields struct {
	Id   string `json:"id"`
	Self string `json:"self"`
}

type IssueLink struct {
	BaseFields
	Type         map[string]string `json:"type"`
	InwardIssue  Issue             `json:"inwardIssue"`
	OutwardIssue Issue             `json:"outwardIssue"`
}

type Issue struct {
	BaseFields
	Key    string      `json:"key"`
	Fields IssueFields `json:fields"`
}

type IssueFields struct {
	Summary        string             `json:"summary"`
	Progress       IssueFieldProgress `json:"progress"`
	IssueType      IssueType          `json:"issuetype"`
	ResolutionDate interface{}        `json:"resolutiondate"`
	Timespent      interface{}        `json:"timespent"`
	Creator        IssueFieldCreator  `json:"creator"`
	Created        string             `json:"created"`
	Updated        string             `json:"updated"`
	Labels         []string           `json:"labels"`
	Assignee       IssueFieldCreator  `json:"assignee"`
	Description    interface{}        `json:"description"`
	IssueLinks     []IssueLink        `json:"issueLinks"`
	Status         IssueStatus        `json:"status"`
}

type IssueFieldProgress struct {
	Progress int `json:"progress"`
	Total    int `json:"total"`
}

type IssueFieldCreator struct {
	Self         string            `json:"self"`
	Name         string            `json:"name"`
	EmailAddress string            `json:"emailAddress"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	DisplayName  string            `json:"displayName"`
	Active       bool              `json:"active"`
}

type IssueType struct {
	BaseFields
	Description string `json:"description"`
	IconUrl     string `json:"iconURL"`
	Name        string `json:"name"`
	Subtask     bool   `json:"subtask"`
}

type IssueStatus struct {
	BaseFields
	Name string `json:"name"`
}

//New will create a new instance of Jira Rest API
func New(credentials config.UserCredential) *RestAPI {
	return &RestAPI{credentials: credentials}
	//var api = new(RestAPI)
	//api.credentials = credentials
	//return api
}

// RestAPI wraps some basic rest calls to work with Jira
type RestAPI struct {
	credentials config.UserCredential
}

// GetJiraRelease returns Epic information
func (api *RestAPI) GetRelease(epicID string) ([]ReleaseItem, error) {
	results := []ReleaseItem{}
	issue, err := api.getIssue(epicID)
	if err != nil {
		return results, err
	}

	scanner := bufio.NewScanner(strings.NewReader(issue.Fields.Description.(string)))
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		if strings.Contains(line, "/app#/projects") {
			parts := strings.Split(line, "/")
			results = append(results, ReleaseItem{Project: parts[5], Version: parts[7]})
		}
	}
	return results, nil
}

func (api *RestAPI) CreateEpicNew(project string, summary string, projectItems []ProjectItem) (*Issue, error) {
	desc := convertToDescription(projectItems)
	i := issue{summary: summary, description: desc, project: project1{key: project}, issuetype: issuetype{name: "epic"}}
	b, _ := json.Marshal(i)
	issue, err := api.createIssue(bytes.NewReader(b))
	return issue, err
}

//CreateEpic will create a new release epic in Jira
func (api *RestAPI) CreateEpic(project string, summary string, releaseItems []ReleaseItem) (*Issue, error) {
	i := issue{summary: summary, project: project1{key: project}, issuetype: issuetype{name: "epic"}}
	b, _ := json.Marshal(i)
	issue, err := api.createIssue(bytes.NewReader(b))
	return issue, err
}

func convertToDescription(pi []ProjectItem) string {
	desc := "..."
	for _, r := range pi {
		desc += fmt.Sprintf("%s|%s\r\n", r.Project, r.Branch)
	}
	return desc
}

func (api *RestAPI) DeleteIssue(issue *Issue) error {
	url := fmt.Sprintf("%s/issue/%s", api.credentials.URL, issue.Key)
	code, body := api.execRequest("DELETE", url, nil)
	if code != http.StatusNoContent {
		return handleJiraError(body)
	}
	return nil
}

func (api *RestAPI) createIssue(params io.Reader) (*Issue, error) {
	url := fmt.Sprintf("%s/issue", api.credentials.URL)
	code, body := api.execRequest("POST", url, params)
	if code == http.StatusCreated {
		response := make(map[string]string)
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return api.getIssue(response["key"])
	}
	return nil, handleJiraError(body)
}

func (api *RestAPI) getIssue(issueKey string) (*Issue, error) {
	url := fmt.Sprintf("%s/issue/%s", api.credentials.URL, issueKey)
	code, body := api.execRequest("GET", url, nil)
	if code == http.StatusOK {
		var issue Issue
		err := json.Unmarshal(body, &issue)
		if err != nil {
			return nil, err
		}
		return &issue, nil
	}
	return nil, handleJiraError(body)
}

// ReleaseItem is a line item in the Jira Release (Epic) that should be deployed
type ReleaseItem struct {
	Project string
	Version string
}

// ProjectItem represents a project and git branch
type ProjectItem struct {
	Project string
	Branch  string
}

func (api *RestAPI) execRequest(requestType, requestUrl string, data io.Reader) (int, []byte) {

	client := &http.Client{}
	req, err := http.NewRequest(requestType, requestUrl, data)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(api.credentials.Username, api.credentials.Password)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode, body
}

func handleJiraError(body []byte) error {
	errorAnswer := ApiError{}
	err := json.Unmarshal(body, &errorAnswer)
	if err != nil {
		return err
	}
	return errors.New(errorAnswer.String())
}

type project1 struct {
	key string
}

type issuetype struct {
	name string
}

type issue struct {
	summary     string
	description string
	project     project1
	issuetype   issuetype
}

/*

{
    "fields": {
       "project":
       {
          "key": "TEST"
       },
       "summary": "REST ye merry gentlemen.",
       "description": "Creating of an issue using project keys and issue type names using the REST API",
       "issuetype": {
          "name": "Bug"
       }
   }
}
*/
