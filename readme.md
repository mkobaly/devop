### This is a work in progress...nothing to see here..yet

## Workflow

- Create build file with Project [branch] listings
- Teamcity kicks off builds asnyc (write tasks to file to check status)
- Poll TC results by reading file and once all complete write out Release file  
- release file consists of Project and version ??
- Create releases in Octopus with release notes
- If selected Create Epic in Jira
- Deploy Epic/Release to selected environment

## Command line thoughts
devop
    - verify epicId ? (very jira specific and not part of devOps but useful)
    - deploy environment [epicId/projectId] - deploy set of projects to the given environment
    - status taskId (displays status of Octopus task)
    - build


build (kick off build on CI server)
deploy (deploy release - this can be one or many projects)
create release 
monitor status

APPNAME VERB NOUN --ADJECTIVE. or APPNAME COMMAND ARG --FLAG
- devop build buildId --branch=master
- devop build --buildFile (buildId and branch)

- devop deploy [jira Epic | projectId] --env= --releaseFile=
- devop deploy --jiraEpic  --projectId --version= --env= --releaseFile=(project/version)

Kick off build and create releases for each project specified in list
- devop assemble --file= --epic 

- devlop verify --jiraEpic --projectId --releaseFile


buildFIle
buildID [branch]
buildID [branch]
----->
Teamcity Release
- buildResult
----->
Octopus
- create Release (ProjectId, version, release notes)
----->
Jira - Create Epic (optional)
----->
Octopus 
- Deploy release



