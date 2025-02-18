package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"path/filepath"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type FileType string

const (
	FileTypeIssueTemplate       FileType = "ISSUE_TEMPLATE"
	FileTypeIssueConfig         FileType = "ISSUE_CONFIG"
	FileTypePullRequestTemplate FileType = "PR_TEMPLATE"
)

var (
	issueTemplateFiles = []string{
		"ISSUE_TEMPLATE.md",
		"ISSUE_TEMPLATE.yaml",
		"ISSUE_TEMPLATE.yml",
		"issue_template.md",
		"issue_template.yaml",
		"issue_template.yml",
		".gitea/ISSUE_TEMPLATE.md",
		".gitea/ISSUE_TEMPLATE.yaml",
		".gitea/ISSUE_TEMPLATE.yml",
		".gitea/issue_template.md",
		".gitea/issue_template.yaml",
		".gitea/issue_template.yml",
		".github/ISSUE_TEMPLATE.md",
		".github/ISSUE_TEMPLATE.yaml",
		".github/ISSUE_TEMPLATE.yml",
		".github/issue_template.md",
		".github/issue_template.yaml",
		".github/issue_template.yml",
	}

	issueConfigFiles = []string{
		".gitea/ISSUE_TEMPLATE/config.yaml",
		".gitea/ISSUE_TEMPLATE/config.yml",
		".gitea/issue_template/config.yaml",
		".gitea/issue_template/config.yml",
		"github/ISSUE_TEMPLATE/config.yaml",
		"github/ISSUE_TEMPLATE/config.yml",
		"github/issue_template/config.yaml",
		"github/issue_template/config.yml",
	}

	issueTemplateDirs = []string{
		"ISSUE_TEMPLATE",
		"issue_template",
		".gitea/ISSUE_TEMPLATE",
		".gitea/issue_template",
		".github/ISSUE_TEMPLATE",
		".github/issue_template",
		".gitlab/ISSUE_TEMPLATE",
		".gitlab/issue_template",
	}

	prTemplateFiles = []string{
		"PULL_REQUEST_TEMPLATE.md",
		"PULL_REQUEST_TEMPLATE.yaml",
		"PULL_REQUEST_TEMPLATE.yml",
		"pull_request_template.md",
		"pull_request_template.yaml",
		"pull_request_template.yml",
		".gitea/PULL_REQUEST_TEMPLATE.md",
		".gitea/PULL_REQUEST_TEMPLATE.yaml",
		".gitea/PULL_REQUEST_TEMPLATE.yml",
		".gitea/pull_request_template.md",
		".gitea/pull_request_template.yaml",
		".gitea/pull_request_template.yml",
		".github/PULL_REQUEST_TEMPLATE.md",
		".github/PULL_REQUEST_TEMPLATE.yaml",
		"github/PULL_REQUEST_TEMPLATE.yml",
		"github/pull_request_template.md",
		"github/pull_request_template.yaml",
		"github/pull_request_template.yml",
	}
)

type TemplateFile struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

func (t TemplateFile) MarshalYAML() (interface{}, error) {
	content := t.Content
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return struct {
		Path    string     `yaml:"path"`
		Content *yaml.Node `yaml:"content"`
	}{
		Path: t.Path,
		Content: &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: content,
			Style: yaml.LiteralStyle,
		},
	}, nil
}

type TemplatesConfig struct {
	IssueTemplates []TemplateFile `yaml:"issue_templates,omitempty"`
	IssueConfigs   []TemplateFile `yaml:"issue_configs,omitempty"`
	PRTemplates    []TemplateFile `yaml:"pr_templates,omitempty"`
}

type TemplatesHandler struct{}

func (h *TemplatesHandler) Name() string {
	return "templates"
}

func (h *TemplatesHandler) Path() string {
	return DefaultTemplatesFile
}

func readTemplates(path string) (TemplatesConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return TemplatesConfig{}, fmt.Errorf("failed to read templates file: %w", err)
	}
	var config TemplatesConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return TemplatesConfig{}, fmt.Errorf("failed to parse templates file: %w", err)
	}
	return config, nil
}

func (h *TemplatesHandler) Load(path string) (interface{}, error) {
	return readTemplates(path)
}

func (h *TemplatesHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
	var config TemplatesConfig

	repository, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	defaultBranch := repository.DefaultBranch

	for _, path := range prTemplateFiles {
		content, resp, err := client.GetFile(owner, repo, defaultBranch, path)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				continue
			}
			return nil, fmt.Errorf("error fetching PR template '%s': %w", path, err)
		}
		config.PRTemplates = append(config.PRTemplates, TemplateFile{
			Path:    path,
			Content: string(content),
		})
	}

	for _, path := range issueConfigFiles {
		content, resp, err := client.GetFile(owner, repo, defaultBranch, path)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				continue
			}
			return nil, fmt.Errorf("error fetching issue config '%s': %w", path, err)
		}
		config.IssueConfigs = append(config.IssueConfigs, TemplateFile{
			Path:    path,
			Content: string(content),
		})
	}

	for _, path := range issueTemplateFiles {
		content, resp, err := client.GetFile(owner, repo, defaultBranch, path)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				continue
			}
			return nil, fmt.Errorf("error fetching issue template '%s': %w", path, err)
		}
		config.IssueTemplates = append(config.IssueTemplates, TemplateFile{
			Path:    path,
			Content: string(content),
		})
	}

	for _, dir := range issueTemplateDirs {
		entries, _, err := client.ListContents(owner, repo, defaultBranch, dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.Type != "file" {
				continue
			}

			if entry.Name == "config.yml" || entry.Name == "config.yaml" {
				continue
			}

			ext := strings.ToLower(filepath.Ext(entry.Name))
			if ext != ".md" && ext != ".yaml" && ext != ".yml" {
				continue
			}

			if entry.Content == nil {
				content, _, err := client.GetFile(owner, repo, defaultBranch, entry.Path)
				if err != nil {
					continue
				}
				config.IssueTemplates = append(config.IssueTemplates, TemplateFile{
					Path:    entry.Path,
					Content: string(content),
				})
			} else {
				config.IssueTemplates = append(config.IssueTemplates, TemplateFile{
					Path:    entry.Path,
					Content: *entry.Content,
				})
			}
		}
	}

	return config, nil
}

type ChangeFilesOptions struct {
	Author    *gitea.Identity          `json:"author,omitempty"`
	Branch    string                   `json:"branch"`
	Committer *gitea.Identity          `json:"committer,omitempty"`
	Dates     *gitea.CommitDateOptions `json:"dates,omitempty"`
	Files     []ChangeFileOperation    `json:"files"`
	Message   string                   `json:"message"`
	NewBranch string                   `json:"new_branch,omitempty"`
	Signoff   bool                     `json:"signoff,omitempty"`
}

type ChangeFileOperation struct {
	Content   string            `json:"content"`
	FromPath  string            `json:"from_path"`
	Operation FileOperationType `json:"operation"`
	Path      string            `json:"path"`
	SHA       string            `json:"sha"`
}

type FileOperationType string

const (
	FileOperationTypeCreate FileOperationType = "create"
	FileOperationTypeUpdate FileOperationType = "update"
	FileOperationTypeDelete FileOperationType = "delete"
)

func (h *TemplatesHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
	templatesConfig, ok := data.(TemplatesConfig)
	if !ok {
		return fmt.Errorf("invalid data type for TemplatesHandler")
	}

	allFiles := make(map[string]string)
	for _, prTemplate := range templatesConfig.PRTemplates {
		allFiles[prTemplate.Path] = prTemplate.Content
	}

	for _, issueTemplate := range templatesConfig.IssueTemplates {
		allFiles[issueTemplate.Path] = issueTemplate.Content
	}

	for _, issueConfig := range templatesConfig.IssueConfigs {
		allFiles[issueConfig.Path] = issueConfig.Content
	}

	toUpdate := make([]ChangeFileOperation, 0)
	toCreate := make([]ChangeFileOperation, 0)

	r, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	baseBranch := r.DefaultBranch
	existingUpdateBranch, _, err := client.GetRepoBranch(owner, repo, DefaultTemplatesUpdateBranchName)
	if err == nil && existingUpdateBranch != nil {
		baseBranch = existingUpdateBranch.Name
	}

	for path, content := range allFiles {
		existingFile, resp, err := client.GetContents(owner, repo, baseBranch, path)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {

				toCreate = append(toCreate, ChangeFileOperation{
					Path:      path,
					Content:   base64.StdEncoding.EncodeToString([]byte(content)),
					Operation: FileOperationTypeCreate,
				})
				continue
			}
			return fmt.Errorf("failed to get content for '%s': %w", path, err)
		}

		if existingFile.Content != nil {
			decodedContent, err := base64.StdEncoding.DecodeString(*existingFile.Content)
			if err != nil {
				return fmt.Errorf("failed to decode existing file '%s': %w", path, err)
			}
			if string(decodedContent) != content {
				toUpdate = append(toUpdate, ChangeFileOperation{
					Path:      path,
					Content:   base64.StdEncoding.EncodeToString([]byte(content)),
					Operation: FileOperationTypeUpdate,
					SHA:       existingFile.SHA,
				})
			}
		}
	}

	allOps := append(toUpdate, toCreate...)
	if len(allOps) == 0 {
		return nil
	}

	opts := ChangeFilesOptions{
		Message:   DefaultTemplatesUpdateCommitMessage,
		Files:     allOps,
		NewBranch: DefaultTemplatesUpdateBranchName,
		Branch:    baseBranch,
	}

	jsonData, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents", cfg.GiteaURL, owner, repo)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", cfg.GiteaToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf(
				"failed to update files, status %d (also failed to read response body: %v)",
				response.StatusCode, err,
			)
		}
		return fmt.Errorf("failed to update files: %s", string(body))
	}

	_, _, err = client.CreatePullRequest(owner, repo, gitea.CreatePullRequestOption{
		Title: DefaultTemplatesUpdateCommitMessage,
		Head:  DefaultTemplatesUpdateBranchName,
		Body:  DefaultTemplatesUpdatePRDescription,
		Base:  r.DefaultBranch,
	})
	if err != nil && !strings.Contains(err.Error(), "pull request already exists") {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	return nil
}

func (h *TemplatesHandler) Enabled() bool {
	return true
}
