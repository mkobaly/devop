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
	"errors"

	"github.com/fatih/color"
	"github.com/mkobaly/devop/config"
	"github.com/mkobaly/devop/jira"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify what projects are part of a release",
	Long: `Verify what projects are part of a release. You can either do this
against a Jira issue if you are using those to track your releases or against a release file

Ex release file 
TODO:

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.NewConfig(viper.ConfigFileUsed())
		jiraAPI := jira.New(config.Jira)
		//tcb := build.TeamcityBuilder{Credentials: config.Teamcity}
		epicID, _ := cmd.Flags().GetString("epicID")
		releaseFile, _ := cmd.Flags().GetString("releaseFile")

		if epicID != "" {
			releases, err := jiraAPI.GetRelease(epicID)
			if err != nil {
				return err
			}
			color.Green("The following applications are part of this release")
			color.Green("--------------------------------------------------------")
			for _, r := range releases {
				color.Green("%s - %s", r.Project, r.Version)
			}
		}

		if releaseFile != "" {

		}
		// TODO: Work your own magic here
		return errors.New("Either epicID or releaseFile are required")
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().StringP("epicID", "e", "", "Verify Jira Epic (release) to see what packages are part of it")
	verifyCmd.Flags().StringP("releaseFile", "f", "", "Release file to verify")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
