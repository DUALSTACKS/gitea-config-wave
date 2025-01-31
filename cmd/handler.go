package cmd

import "code.gitea.io/sdk/gitea"

type ConfigHandler interface {
	Name() string
	Path() string
	Enabled() bool
	Pull(client *gitea.Client, owner, repo string) (interface{}, error)
	Push(client *gitea.Client, owner, repo string, data interface{}) error
	Load(path string) (interface{}, error)
}
