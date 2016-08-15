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
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mkobaly/devop/config"
	"github.com/mkobaly/devop/jira"
	"github.com/mkobaly/devop/octopus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy environment",
	Short: "Deploy a set of project(s) from Octopus Deploy",
	Long: `Will deploy a set of projects(s) stored in Octopus
Deploy to the selected Environment. It will monitor each
package for their final results and display them in the console.

Examples:

devop deploy staging -p myProject
devop deploy production -e abc-123
devop deploy production -e abc-123 -l deploy.log

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Environment not specified")
		}

		config := config.NewConfig(viper.ConfigFileUsed())
		octo := octopus.New(config.Octopus.URL, config.Octopus.Webapikey)
		jiraAPI := jira.New(config.Jira)
		taskChan := make(chan octopus.TaskResult)
		env, err := validateEnvironment(args[0], octo)
		if err != nil {
			return err
		}

		var releaseItems []jira.ReleaseItem
		epic, _ := cmd.Flags().GetString("epic")
		project, _ := cmd.Flags().GetString("project")
		deployFile, _ := cmd.Flags().GetString("deployFile")
		version, _ := cmd.Flags().GetString("version")

		if epic == "" && project == "" && deployFile == "" {
			return errors.New("An epic or project or deploy file must be specified")
		}

		if project != "" {
			if version == "" {
				return errors.New("Version must be specified when deploying indivdual project")
			}
			releaseItems = append(releaseItems, jira.ReleaseItem{Project: project, Version: version})
		} else if epic != "" {
			releaseItems, err = jiraAPI.GetRelease(epic)
			if err != nil {
				return err
			}
		} else if deployFile != "" {
			releaseItems, err = parseDeployFile(deployFile)
		}

		tasks, err := deploy(releaseItems, env, octo)

		for _, task := range tasks {
			//kick off goroutine to watch all projects getting deployed
			go watchTaskForResult(task, octo, taskChan)
		}

		for i := 0; i < len(tasks); i++ {
			reportTaskResult(<-taskChan)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("epic", "e", "", "Jira epic to deploy")
	deployCmd.Flags().StringP("deployFile", "d", "", "Deploy file containing projects to deploy")
	deployCmd.Flags().StringP("project", "p", "", "Individual octopus project to deploy")
	deployCmd.Flags().StringP("version", "v", "", "Version for individual project to deploy")
	deployCmd.Flags().StringP("logFile", "l", "", "Log deployment results to file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

//validateEnvironment ensures the user passed in a valid Octopus
//environment and that we are not deploying to production as that
// is not allowed
func validateEnvironment(env string, octo *octopus.Octo) (octopus.Environment, error) {
	var e octopus.Environment
	if env == "production" {
		return e, errors.New("Deploying to prodution is not allowed due to ISO requirements")
	}
	envs, err := octo.GetEnvironments()
	if err != nil {
		panic(err.Error())
	}
	envString := []string{}
	for _, x := range envs {
		if strings.ToLower(x.Name) == strings.ToLower(env) {
			return x, nil
		}
		if strings.ToLower(x.Name) != "production" {
			envString = append(envString, x.Name)
		}
	}
	return e, errors.New("Unknown environment " + env +
		". Valid values are:\n\t" + strings.Join(envString, "\n\t"))
}

//watchTaskForResult will poll Octopus for the result of a deployments
//and once its completed will inform on resultChan
func watchTaskForResult(t octopus.TaskID, o *octopus.Octo, resultChan chan octopus.TaskResult) {
	//time.Sleep(time.Second * 5)
	for {
		result, err := o.GetTaskResult(t.TaskID)
		if err == nil {
			if result.IsCompleted {
				resultChan <- result
				return
			}
		}
		time.Sleep(time.Second * 2)
	}
}

//reportTaskResult will display the results of a project
//getting deployed color coded
func reportTaskResult(t octopus.TaskResult) {
	if t.FinishedSuccessfully {
		color.Green("%s - %s. Duration:%s ", t.State, t.Description, t.Duration)
	} else {
		color.Red("%s - %s Error: %s", t.State, t.Description, t.ErrorMessage)
	}
}

//deployEpic will deploy all releases to the specified environment
func deploy(releaseItems []jira.ReleaseItem, env octopus.Environment, octo *octopus.Octo) ([]octopus.TaskID, error) {
	tasks := []octopus.TaskID{}
	//releaseItems, err := GetJiraRelease(epic, config)
	// if err != nil {
	// 	return tasks, err
	// 	//color.Red(err.Error())
	// }

	for _, r := range releaseItems {
		projectID, _ := octo.GetProjectID(r.Project)
		releaseID, _ := octo.GetReleaseID(projectID, r.Version)
		ID, _ := octo.Deploy(releaseID, env.ID) //test(releaseID)
		color.Green("Deploying %s. TaskId: %s", r.Project, ID.TaskID)
		tasks = append(tasks, ID)
	}
	return tasks, nil
}

//Will parse a deploy file that lists out each project version that needs to be deployed
func parseDeployFile(path string) ([]jira.ReleaseItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ri []jira.ReleaseItem
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 2 {
			return nil, errors.New("you must specify project and version")
		}
		ri = append(ri, jira.ReleaseItem{Project: parts[0], Version: parts[1]})
	}
	return ri, nil
}
