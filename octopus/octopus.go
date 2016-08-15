package octopus

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Internal json results returned from Octopus REST API
type id struct {
	ID string
}

type result struct {
	ItemType     string
	IsStale      bool
	TotalResults int
	ItemsPerPage int
	Items        []Environment
}

//Octo represents Octopus Deploy
type Octo struct {
	url    string
	apiKey string
}

//TaskID represents a octopus task
type TaskID struct {
	TaskID string
}

//TaskResult is an Octopus Deploy deployment result status
type TaskResult struct {
	ID                   string
	Name                 string
	Description          string
	State                string
	Duration             string
	IsCompleted          bool
	FinishedSuccessfully bool
	ErrorMessage         string
}

//Environment defined in Octopus Deploy
type Environment struct {
	ID   string
	Name string
}

//New will create a new instance of Octo
func New(url string, apiKey string) *Octo {
	var octo = new(Octo)
	octo.url = url
	octo.apiKey = apiKey
	return octo
}

//GetEnvironments will return all of the environments defined in Octopus Deploy
func (o *Octo) GetEnvironments() ([]Environment, error) {
	url := o.url + "/environments"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		panic(err.Error())
	}

	var r result
	er := decode(resp, &r)
	return r.Items, er
}

//GetProjectID will return the projectId for a given project name
func (o *Octo) GetProjectID(p string) (string, error) {
	url := o.url + "/projects/" + p

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		panic(err.Error())
	}

	var result id
	er := decode(resp, &result)
	return result.ID, er
}

//GetReleaseID returns a release id for a given projectId and release
//projectId: projects-xxx
//release: 3.3.4.0
func (o *Octo) GetReleaseID(projectID string, release string) (string, error) {
	url := o.url + "/projects/" + projectID + "/releases/" + release

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		panic(err.Error())
	}
	var result id
	er := decode(resp, &result)

	return result.ID, er
}

//GetTaskResult will return the status of a given task (deployment)
func (o *Octo) GetTaskResult(taskID string) (TaskResult, error) {
	url := o.url + "/tasks/" + taskID

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var result TaskResult
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return result, err
	}

	err = decode(resp, &result)
	return result, err
}

//Deploy an project specific release to the given environment
func (o *Octo) Deploy(releaseID string, environmentID string) (TaskID, error) {
	url := o.url + "/deployments/"

	data := struct {
		ReleaseID     string
		EnvironmentID string
	}{
		releaseID,
		environmentID,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(data)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, b)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		panic(err.Error())
	}
	var result TaskID
	er := decode(resp, &result)

	return result, er
}

//CreateRelease will create a new release for a given project at the given version
func (o *Octo) CreateRelease(projectID string, version string, releaseNotes string) error {
	url := o.url + "/releases/"
	data := struct {
		ProjectID    string
		Version      string
		ReleaseNotes string
	}{
		projectID,
		version,
		releaseNotes,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(data)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, b)
	req.Header.Set("X-Octopus-ApiKey", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	return err

}

func decode(r *http.Response, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}
