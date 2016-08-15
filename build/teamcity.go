package build

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/mkobaly/devop/config"
	"github.com/mkobaly/teamcity"
)

//BuildInfo represents a build and branch in Teamcity
type BuildInfo struct {
	BuildID string
	Branch  string
}

type Builder interface {
	//Build(id string, branch string) error
	Build() error
}

type TeamcityBuilder struct {
	Credentials config.UserCredential
	BuildID     string
	Branch      string
	BuildResult *teamcity.Build
}

//Build will kick off a TeamCity build
func (b *TeamcityBuilder) Build() error {
	client := teamcity.New(b.Credentials.URL, b.Credentials.Username, b.Credentials.Password)
	x, err := client.QueueBuild(b.BuildID, b.Branch, nil)
	if err != nil {
		return err
	}
	b.BuildResult = x
	return nil
}

func (b *TeamcityBuilder) GetBuildResult() string {
	r, _ := json.MarshalIndent(b.BuildResult, "", "\t")
	//json := string(r)
	return string(r)
	//return fmt.Sprintf("Kicking off build for %s using branch %s\nUrl to build result:  %s\n",
	//	b.BuildResult.BuildType.ProjectName, b.BuildResult.BranchName, b.BuildResult.WebURL)
	//color.Green("Url to build result:  %s\n", tcb.BuildResult.WebURL)
}

//Will parse a build file that lists out each buildID [branch] that needs to be built
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
		builds = append(builds, BuildInfo{BuildID: parts[0], Branch: branch})
	}
	return builds, nil
}
