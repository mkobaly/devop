// Copyright © 2016 Michael Kobaly mkobaly@gmail.com
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
	"time"

	"strings"

	"github.com/fatih/color"
	"github.com/mkobaly/devop/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

import tc "github.com/mkobaly/devop/teamcity"
import teamcity "github.com/mkobaly/teamcity"

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [buildId]",
	Short: "Kick off a Teamcity build",
	Long:  "This will start a Teamcity build of one or more projects",
	Example: strings.Join([]string{
		"- devop build projectA                 Kick off build of projectA",
		"- devop build projectA -b abc          Kick off build of projectA using branch abc",
		"- devop build -f build.txt             Kick off build of all projects listed in build.txt",
		"- devop build projectA -l results.log  Kick off build of projectA and log build results to a file",
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {

		config := config.NewConfig(viper.ConfigFileUsed())
		tcb := tc.New(config.Teamcity)
		//tcb := tc.Builder{Credentials: config.Teamcity}
		logFile, _ := cmd.Flags().GetString("logFile")

		if logFile != "" {
			_ = os.Remove(logFile)
		}

		buildInfo := []tc.BuildInfo{}

		if len(args) == 0 {
			buildFile, _ := cmd.Flags().GetString("buildFile")
			if buildFile == "" {
				return errors.New("You must provide a buildId or set the buildFile flag")
			}
			builds, err := tc.ParseBuildFile(buildFile)
			if err != nil {
				return err
			}
			for _, b := range builds {
				buildInfo = append(buildInfo, b)
				//tcBuild(&tcb, b.BuildID, b.Branch, logFile)
			}

		} else { //single build
			branch, _ := cmd.Flags().GetString("branch")
			buildID := args[0]
			buildInfo = append(buildInfo, tc.BuildInfo{BuildConfigID: buildID, Branch: branch})
			//tcBuild(&tcb, buildID, branch, logFile)
		}

		buildChan := make(chan teamcity.Build)
		for _, bi := range buildInfo {
			tcb.SetBuildInfo(bi)
			if err := tcb.Build(); err != nil {
				return err
			}
			//tcBuild(&tcb, &bi, logFile)
			go watchForFinishedBuild(tcb, buildChan)
		}

		for i := 0; i < len(buildInfo); i++ {
			reportResult(<-buildChan)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("branch", "b", "", "Branch to build (Default branch used if blank)")
	buildCmd.Flags().StringP("buildFile", "f", "", "Build file listing out each project and branch to build")
	buildCmd.Flags().StringP("logFile", "l", "", "Log build results to file")

	//buildCmd.Flags().StringP("projectId", "p", "", "Project to build")
}

func watchForFinishedBuild(b *tc.Builder, c chan teamcity.Build) error {
	for {
		time.Sleep(time.Second * 2)
		err := b.VerifyBuildStatus()
		if err != nil {
			return err
		}
		fmt.Print(".")
		if b.BuildResult.State == "finished" {
			c <- *b.BuildResult
			return nil
		}
	}
}

func reportResult(b teamcity.Build) {
	if b.Status == "SUCCESS" {
		color.Green("\n%s %s: %s ", b.BuildTypeID, b.State, b.Status)
	} else {
		color.Red("\n%s %s: %s", b.BuildTypeID, b.State, b.Status)
	}
}

// func tcBuild(b *tc.Builder, bi *tc.BuildInfo, logFile string) error {
// 	if bi.Branch == "" {
// 		bi.Branch = ""
// 	}
// 	b.BuildInfo = bi

// 	if err := b.Build(); err != nil {
// 		return err
// 	}
// 	branchName := bi.Branch
// 	if branchName == "" {
// 		branchName = "[Default]"
// 	}
// 	color.Green("Kicking off build for %s using branch %s\n",
// 		b.BuildResult.BuildType.ProjectName, branchName)
// 	if logFile != "" {
// 		writeToLog(logFile, b.BuildResultToJson())
// 	}
// 	return nil
// }

func writeToLog(path string, content string) {
	fileHandle, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666)
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, content)
	writer.Flush()
}
