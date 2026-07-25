package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const repoOwner = "c3lang"
const repoName = "c3c"
const apiBase = "https://api.github.com"

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func ListReleases(page int) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100&page=%d", apiBase, repoOwner, repoName, page)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}
	return releases, nil
}

func GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBase, repoOwner, repoName)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}
	return &release, nil
}

func DownloadURL(tag, assetName string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", repoOwner, repoName, tag, assetName)
}

func ListAllTags() ([]string, error) {
	var tags []string
	page := 1
	for {
		releases, err := ListReleases(page)
		if err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			break
		}
		for _, r := range releases {
			tags = append(tags, r.TagName)
		}
		page++
	}
	return tags, nil
}
