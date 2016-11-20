package teamcity

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	// "fmt"
	// "io/ioutil"
	"net/http"

	"github.com/mkobaly/devop/config"
	"github.com/mkobaly/teamcity"
)

//BuildInfo represents a Build Configuration and branch in Teamcity
type BuildInfo struct {
	BuildConfigID string
	Branch        string
}

type IBuilder interface {
	//Build(id string, branch string) error
	Build() error
	Done() <-chan struct{}
	Result() interface{}
}

type Builder struct {
	Credentials config.UserCredential
	BuildInfo   BuildInfo
	// buildID     string
	// branch      string
	client      *teamcity.Client
	BuildResult *teamcity.Build
}

//New will create a new teamcity Builder
func New(creds config.UserCredential) *Builder {
	var b = new(Builder)
	b.Credentials = creds
	//b.BuildInfo = buildInfo
	//b.branch = branch
	//b.buildID = buildID
	b.client = teamcity.New(creds.URL, creds.Username, creds.Password)
	return b
}

func (b *Builder) SetBuildInfo(bi BuildInfo) error {
	b.BuildInfo = bi
	b.BuildResult = new(teamcity.Build)
	return nil
}

//Build will kick off a TeamCity build
func (b *Builder) Build() error {
	if (BuildInfo{}) == b.BuildInfo {
		return errors.New("Build Info not set yet so unable to build")
	}
	//client := teamcity.New(b.Credentials.URL, b.Credentials.Username, b.Credentials.Password)
	x, err := b.client.QueueBuild(b.BuildInfo.BuildConfigID, b.BuildInfo.Branch, nil)
	if err != nil {
		return err
	}
	b.BuildResult = x
	return nil
}

func (b *Builder) BuildResultToJson() string {
	r, _ := json.MarshalIndent(b.BuildResult, "", "\t")
	return string(r)
}

// //GetBuild will return the current state of the build
func (b *Builder) GetBuild() error {
	br, err := b.client.GetBuild(strconv.FormatInt(b.BuildResult.ID, 10))
	b.BuildResult = br
	return err
}

//VerifyBuildStatus will return the current state of the build
func (b *Builder) VerifyBuildStatus() error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", b.Credentials.URL+b.BuildResult.HREF, nil)
	req.SetBasicAuth(b.Credentials.Username, b.Credentials.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		panic(err.Error())
	}
	if err := json.NewDecoder(resp.Body).Decode(b.BuildResult); err != nil {
		return err
	}
	return nil
}

//GetArtifactVersion will return the version number of the build artifact
func (b *Builder) GetArtifactVersion() (string, error) {
	client := teamcity.New(b.Credentials.URL, b.Credentials.Username, b.Credentials.Password)
	version, err := client.GetArtifact(b.BuildResult.ID)
	if err != nil {
		var s = strings.Replace(version.Name, ".zip", "", 1)
		var parts = strings.Split(s, ".v")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}
	return "", err
}

//ParseBuildFile Will parse a build file that lists out each buildID [branch] that needs to be built
//buildID is required but branch is not. It will default to master. BuildID and branch name
//need to be separated by a space
func ParseBuildFile(path string) ([]BuildInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var builds []BuildInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		branch := ""
		if len(parts) > 1 {
			branch = parts[1]
		}
		builds = append(builds, BuildInfo{BuildConfigID: parts[0], Branch: branch})
	}
	return builds, nil
}

//GetBuilds will list out all available builds on TeamCity
func (b *Builder) GetBuilds() ([]*teamcity.BuildType, error) {
	//client := teamcity.New(b.Credentials.URL, b.Credentials.Username, b.Credentials.Password)
	builds, err := b.client.GetBuildTypes()
	return builds, err
}
