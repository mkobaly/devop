// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mkobaly/devop/build"
	"github.com/mkobaly/devop/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [buildId]",
	Short: "Kick off the specified build on your Build Server",
	Long: `This will start a build of a specific project on your build server. 
For example:

devop build my-project
devop build my-project -b abc (use branch abc)`,
	RunE: func(cmd *cobra.Command, args []string) error {

		config := config.NewConfig(viper.ConfigFileUsed())
		tcb := build.TeamcityBuilder{Credentials: config.Teamcity}
		logFile, _ := cmd.Flags().GetString("logFile")

		if logFile != "" {
			_ = os.Remove(logFile)
		}

		if len(args) == 0 {
			buildFile, _ := cmd.Flags().GetString("buildFile")
			if buildFile == "" {
				return errors.New("You must provide a buildId or set the buildFile flag")
			}
			builds, err := build.ParseBuildFile(buildFile)
			if err != nil {
				return err
			}
			for _, b := range builds {
				tcBuild(&tcb, b.BuildID, b.Branch, logFile)
			}

		} else { //single build
			branch, _ := cmd.Flags().GetString("branch")
			buildID := args[0]
			tcBuild(&tcb, buildID, branch, logFile)
		}
		return nil
		//fmt.Println("build called")
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("branch", "b", "", "Branch to build (Default branch used if blank)")
	buildCmd.Flags().StringP("buildFile", "f", "", "Build file listing out each project and branch to build")
	buildCmd.Flags().StringP("logFile", "l", "", "Log build results to file")

	//buildCmd.Flags().StringP("projectId", "p", "", "Project to build")
}

func tcBuild(b *build.TeamcityBuilder, buildID string, branch string, logFile string) error {
	b.Branch = branch
	if branch == "" {
		b.Branch = ""
	}
	b.BuildID = buildID
	if err := b.Build(); err != nil {
		return err
	}
	branchName := branch
	if branchName == "" {
		branchName = "[Default]"
	}
	color.Green("Kicking off build for %s using branch %s\nUrl to build result:  %s\n",
		b.BuildResult.BuildType.ProjectName, branchName, b.BuildResult.WebURL)
	if logFile != "" {
		writeToLog(logFile, b.GetBuildResult())
	}
	return nil
}

func writeToLog(path string, content string) {
	fileHandle, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666)
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, content)
	writer.Flush()
}

//Will parse a build file that lists out each buildID [branch] that needs to be built
//buildID is required but branch is not. It will default to master. BuildID and branch name
//need to be separated by a space
// func parseBuildFile(path string) ([]buildInfo, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var builds []buildInfo
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		parts := strings.Split(scanner.Text(), " ")
// 		branch := ""
// 		if len(parts) > 1 {
// 			branch = parts[1]
// 		}
// 		builds = append(builds, buildInfo{buildID: parts[0], branch: branch})
// 	}
// 	return builds, nil
// }
