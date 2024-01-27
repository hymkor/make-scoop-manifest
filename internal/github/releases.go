package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func queryReleases(user, repo string, log io.Writer) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)
	fmt.Fprintln(log, "Get:", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
	tagName            string
}

type Release struct {
	TagName string   `json:"tag_name"`
	Assets  []*Asset `json:"assets"`
}

func GetReleases(name, repo string, log io.Writer) ([]*Release, error) {
	releasesStr, err := queryReleases(name, repo, log)
	if err != nil {
		return nil, fmt.Errorf("getReleases: %w", err)
	}
	var releases []*Release
	if err := json.Unmarshal(releasesStr, &releases); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return releases, nil
}
