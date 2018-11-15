package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/chartmuseum/helm-push/pkg/helm"
	hub "github.com/imroc/helm-push/pkg/tencenthub"
	"github.com/spf13/cobra"
)

var (
	globalUsage = `Helm plugin to push chart package to TencentHub

Examples:

  $ helm push mychart-0.1.0.tgz myhubrepo                     # push .tgz from "helm package"
  $ helm push . myhubrepo                                     # package and push chart directory
  $ helm push . --version="1.2.3" myhubrepo                   # override version in Chart.yaml
  $ helm push . https://hub.tencentyun.com/charts/myhubrepo   # push directly to chart repo URL
`
)

type (
	pushCmd struct {
		chartName    string
		chartVersion string
		repoName     string
		username     string
		password     string
	}
)

func newPushCmd(args []string) *cobra.Command {
	p := &pushCmd{}
	cmd := &cobra.Command{
		Use:          "helm push",
		Short:        "Helm plugin to push chart package to TencentHub",
		Long:         globalUsage,
		SilenceUsage: false,
		Args:         cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("This command needs 2 arguments: name of chart, name of chart repository (or repo URL)")
			}
			p.chartName = args[0]
			p.repoName = args[1]
			p.setFieldsFromEnv()
			return p.push()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&p.chartVersion, "version", "v", "", "Override chart version pre-push")
	f.StringVarP(&p.username, "username", "u", "", "Override HTTP basic auth username [$HELM_REPO_USERNAME]")
	f.StringVarP(&p.password, "password", "p", "", "Override HTTP basic auth password [$HELM_REPO_PASSWORD]")
	f.Parse(args)
	return cmd
}

func (p *pushCmd) setFieldsFromEnv() {
	if v, ok := os.LookupEnv("HELM_REPO_USERNAME"); ok && p.username == "" {
		p.username = v
	}
	if v, ok := os.LookupEnv("HELM_REPO_PASSWORD"); ok && p.password == "" {
		p.password = v
	}
}

func (p *pushCmd) push() error {
	var repo *helm.Repo
	var err error

	// If the argument looks like a URL, just create a temp repo object
	// instead of looking for the entry in the local repository list
	if regexp.MustCompile(`^https?://`).MatchString(p.repoName) {
		repo, err = helm.TempRepoFromURL(p.repoName)
		p.repoName = repo.URL
	} else {
		repo, err = helm.GetRepoByName(p.repoName)
	}

	if err != nil {
		return err
	}

	chart, err := helm.GetChartByName(p.chartName)
	if err != nil {
		return err
	}

	// version override
	if p.chartVersion != "" {
		chart.SetVersion(p.chartVersion)
	}

	// username/password override(s)
	username := repo.Username
	password := repo.Password
	if p.username != "" {
		username = p.username
	}
	if p.password != "" {
		password = p.password
	}

	tmp, err := ioutil.TempDir("", "helm-push-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	chartPackagePath, err := helm.CreateChartPackage(chart, tmp)
	if err != nil {
		return err
	}

	fmt.Printf("Pushing %s to %s...\n", filepath.Base(chartPackagePath), p.repoName)

	client := hub.NewClient(
		hub.URL(repo.URL),
		hub.Username(username),
		hub.Password(password),
	)
	resp, err := client.UploadChartPackage(chartPackagePath, true)
	if err != nil {
		return err
	}
	return handlePushResponse(resp)
}

func handlePushResponse(resp *http.Response) error {
	if resp.StatusCode != 201 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return getTencentHubError(b, resp.StatusCode)
	}
	fmt.Println("Done.")
	return nil
}

func getTencentHubError(b []byte, code int) error {
	return fmt.Errorf("%d: %s", code, b)
}

func main() {
	cmd := newPushCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
