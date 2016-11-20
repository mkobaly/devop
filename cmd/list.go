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
	"strings"

	"github.com/fatih/color"
	"github.com/mkobaly/devop/config"
	"github.com/mkobaly/devop/jira"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tc "github.com/mkobaly/devop/teamcity"
)

// verifyCmd represents the verify command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List out details of teamcity build types or applications part of Jira Epic",
	Long:  "Will list out all of the teamcity build types or list out all of the projects that are part of a Jira Epic",
	Example: strings.Join([]string{
		"- devop list                  List all available build Types for TeamCity",
		"- devop list -e epicId        List projects that are part of Jira Epic",
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.NewConfig(viper.ConfigFileUsed())

		epicID, _ := cmd.Flags().GetString("epicID")
		if epicID == "" {
			return getBuildTypes(config)
		}
		return epicDetails(cmd, config)
	},
}

func getBuildTypes(config *config.Config) error {
	tcb := tc.New(config.Teamcity)
	builds, err := tcb.GetBuilds()

	if err == nil {
		color.Cyan("--------------------------------------------------------")
		color.Cyan("Teamcity Build Types")
		color.Cyan("--------------------------------------------------------")
		for _, r := range builds {
			color.Green("%s", r.ID)
		}
		return nil
	}
	return err
}

func epicDetails(cmd *cobra.Command, config *config.Config) error {
	jiraAPI := jira.New(config.Jira)
	epicID, _ := cmd.Flags().GetString("epicID")

	if epicID != "" {
		releases, err := jiraAPI.GetRelease(epicID)
		if err != nil {
			return err
		}
		color.Cyan("----------------------------------------------------------")
		color.Cyan("The following applications are part of epic: %s", epicID)
		color.Cyan("----------------------------------------------------------")
		for _, r := range releases {
			color.Green("%s - %s", r.Project, r.Version)
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("epicID", "e", "", "Examine Jira Epic (release) to see what packages are part of it")
	//listCmd.Flags().StringP("releaseFile", "f", "", "Release file to verify")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
