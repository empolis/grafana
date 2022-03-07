package build

import "io/ioutil"

func getGitBranch() string {
	v, err := runError("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		v, err = ioutil.ReadFile("git-branch")
		if err != nil {
			return "main"
		}
	}
	return string(v)
}

func getGitSha() string {
	v, err := runError("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		v, err = ioutil.ReadFile("git-sha")
		if err != nil {
			return "unknown-dev"
		}
	}
	return string(v)
}
