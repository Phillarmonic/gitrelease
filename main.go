package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

// Release interface to abstract different providers
type Release interface {
	GetTagName() string
}

// GitHubRelease represents the latest release information from GitHub
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// GetTagName returns the tag name for GitHubRelease
func (gr GitHubRelease) GetTagName() string {
	return gr.TagName
}

// GitLabRelease represents the latest release information from GitLab
type GitLabRelease struct {
	TagName string `json:"tag_name"`
}

// GetTagName returns the tag name for GitLabRelease
func (gl GitLabRelease) GetTagName() string {
	return gl.TagName
}

// BitbucketTag represents a tag in Bitbucket
type BitbucketTag struct {
	Name string `json:"name"`
}

// BitbucketTagsResponse represents the response from Bitbucket tags API
type BitbucketTagsResponse struct {
	Values   []BitbucketTag `json:"values"`
	Pagelen  int            `json:"pagelen"`
	Next     string         `json:"next,omitempty"`
	Previous string         `json:"previous,omitempty"`
}

// GitHubTag represents a single tag in GitHub
type GitHubTag struct {
	Name string `json:"name"`
}

// GitLabTag represents a single tag in GitLab
type GitLabTag struct {
	Name string `json:"name"`
}

// Helper function to URL-encode project path for GitLab
func encodeGitLabProjectPath(path string) string {
	return url.PathEscape(path)
}

func main() {
	// Command-line flags
	repoPtr := flag.String("repo", "", "Repository in the format 'owner/repo' or 'namespace/project' for GitLab")
	providerPtr := flag.String("provider", "github", "Provider: github, gitlab, bitbucket (default: github)")
	versionPtr := flag.String("version", "", "Specify the version prefix (e.g., '8.2') to fetch the latest tag matching that version")
	githubTokenPtr := flag.String("github-token", "", "GitHub Personal Access Token (optional)")
	gitlabTokenPtr := flag.String("gitlab-token", "", "GitLab Personal Access Token (optional)")
	bitbucketTokenPtr := flag.String("bitbucket-token", "", "Bitbucket App Password (optional)")

	flag.Parse()

	if *repoPtr == "" {
		// Define the program version
		const programVersion = "GitRelease 2.2.0"
		fmt.Println(programVersion)
		fmt.Println("🍏🐧 Now for Darwinians and Tuxedos :)")
		fmt.Println("\nUsage:")
		fmt.Println("  gitrelease -repo=<owner/repo or namespace/project> [-provider=<github|gitlab|bitbucket>] [-version=<version_prefix>] [-github-token=<token>] [-gitlab-token=<token>] [-bitbucket-token=<token>]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	provider := strings.ToLower(*providerPtr)
	repo := *repoPtr
	versionPrefix := *versionPtr

	var tag string
	var err error

	client := &http.Client{Timeout: 10 * time.Second}

	switch provider {
	case "github":
		if versionPrefix != "" {
			tag, err = fetchGitHubLatestTagByVersion(repo, client, "", *githubTokenPtr, versionPrefix)
		} else {
			tag, err = fetchGitHubLatestRelease(repo, client, "", *githubTokenPtr)
		}
	case "gitlab":
		if versionPrefix != "" {
			tag, err = fetchGitLabLatestTagByVersion(repo, client, "", *gitlabTokenPtr, versionPrefix)
		} else {
			tag, err = fetchGitLabLatestRelease(repo, client, "", *gitlabTokenPtr)
		}
	case "bitbucket":
		if versionPrefix != "" {
			tag, err = fetchBitbucketLatestTagByVersion(repo, client, "", *bitbucketTokenPtr, versionPrefix)
		} else {
			tag, err = fetchBitbucketLatestTag(repo, client, "", *bitbucketTokenPtr)
		}
	default:
		log.Fatalf("Unsupported provider: %s. Supported providers are github, gitlab, bitbucket.", provider)
	}

	if err != nil {
		log.Fatalf("Error fetching latest release/tag: %v", err)
	}

	fmt.Println(tag)
}

// Fetch the latest release tag from GitHub
func fetchGitHubLatestRelease(repo string, client *http.Client, baseURL string, token string) (string, error) {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/releases/latest", baseURL, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub request: %v", err)
	}

	// Set User-Agent header as required by GitHub API
	req.Header.Set("User-Agent", "gitrelease-cli")

	// Optional: Set Authorization header if token is provided via flag or environment variable
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	} else if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", envToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub release: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		resetUnix, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
		resetTime := time.Unix(resetUnix, 0)
		return "", fmt.Errorf("GitHub rate limit exceeded. Try again at %s", resetTime)
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("GitHub repository or release not found: %s", repo)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GitHub response body: %v", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse GitHub JSON: %v", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found for GitHub repository: %s", repo)
	}

	return release.TagName, nil
}

// Fetch the latest tag from GitHub matching a specific version prefix
func fetchGitHubLatestTagByVersion(repo string, client *http.Client, baseURL string, token string, versionPrefix string) (string, error) {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	// GitHub tags are paginated; fetch all tags and filter
	perPage := 100
	page := 1
	var allTags []GitHubTag

	for {
		url := fmt.Sprintf("%s/repos/%s/tags?per_page=%d&page=%d", baseURL, repo, perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create GitHub tags request: %v", err)
		}

		req.Header.Set("User-Agent", "gitrelease-cli")

		if token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		} else if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", envToken))
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch GitHub tags: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetUnix, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
			resetTime := time.Unix(resetUnix, 0)
			return "", fmt.Errorf("GitHub rate limit exceeded. Try again at %s", resetTime)
		}

		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("GitHub repository not found: %s", repo)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read GitHub tags response body: %v", err)
		}

		var tags []GitHubTag
		if err := json.Unmarshal(body, &tags); err != nil {
			return "", fmt.Errorf("failed to parse GitHub tags JSON: %v", err)
		}

		if len(tags) == 0 {
			break
		}

		allTags = append(allTags, tags...)

		// Check if there are more pages
		if len(tags) < perPage {
			break
		}
		page++
	}

	// Filter tags by version prefix
	var filteredTags []string
	for _, tag := range allTags {
		if strings.HasPrefix(tag.Name, "php-"+versionPrefix) {
			filteredTags = append(filteredTags, tag.Name)
		}
	}

	if len(filteredTags) == 0 {
		return "", fmt.Errorf("no tags found for version prefix: %s", versionPrefix)
	}

	// Sort the filtered tags using semantic versioning
	sortedTags, err := sortTagsByVersion(filteredTags)
	if err != nil {
		return "", fmt.Errorf("failed to sort tags: %v", err)
	}

	return sortedTags[0], nil // Assuming sortedTags is descending
}

// Fetch the latest release tag from GitLab
func fetchGitLabLatestRelease(repo string, client *http.Client, baseURL string, token string) (string, error) {
	// GitLab API expects the project ID or URL-encoded namespace/project
	encodedRepo := encodeGitLabProjectPath(repo)
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	url := fmt.Sprintf("%s/projects/%s/releases", baseURL, encodedRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GitLab request: %v", err)
	}

	// Optional: Set Private-Token header if token is provided via flag or environment variable
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	} else if envToken := os.Getenv("GITLAB_TOKEN"); envToken != "" {
		req.Header.Set("PRIVATE-TOKEN", envToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitLab releases: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting (GitLab uses standard HTTP status codes)
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return "", fmt.Errorf("GitLab rate limit exceeded. Retry after %s seconds", retryAfter)
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("GitLab repository or releases not found: %s", repo)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GitLab response body: %v", err)
	}

	var releases []GitLabRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to parse GitLab JSON: %v", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found for GitLab repository: %s", repo)
	}

	return releases[0].TagName, nil // GitLab returns releases in descending order
}

// Fetch the latest tag from GitLab matching a specific version prefix
func fetchGitLabLatestTagByVersion(repo string, client *http.Client, baseURL string, token string, versionPrefix string) (string, error) {
	encodedRepo := encodeGitLabProjectPath(repo)
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	perPage := 100
	page := 1
	var allTags []GitLabTag

	for {
		url := fmt.Sprintf("%s/projects/%s/repository/tags?per_page=%d&page=%d", baseURL, encodedRepo, perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create GitLab tags request: %v", err)
		}

		if token != "" {
			req.Header.Set("PRIVATE-TOKEN", token)
		} else if envToken := os.Getenv("GITLAB_TOKEN"); envToken != "" {
			req.Header.Set("PRIVATE-TOKEN", envToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch GitLab tags: %v", err)
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			return "", fmt.Errorf("GitLab rate limit exceeded. Retry after %s seconds", retryAfter)
		}

		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("GitLab repository not found: %s", repo)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GitLab API error: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read GitLab tags response body: %v", err)
		}

		var tags []GitLabTag
		if err := json.Unmarshal(body, &tags); err != nil {
			return "", fmt.Errorf("failed to parse GitLab tags JSON: %v", err)
		}

		if len(tags) == 0 {
			break
		}

		allTags = append(allTags, tags...)

		// Check if there are more pages
		if len(tags) < perPage {
			break
		}
		page++
	}

	// Filter tags by version prefix
	var filteredTags []string
	for _, tag := range allTags {
		if strings.HasPrefix(tag.Name, "php-"+versionPrefix) {
			filteredTags = append(filteredTags, tag.Name)
		}
	}

	if len(filteredTags) == 0 {
		return "", fmt.Errorf("no tags found for version prefix: %s", versionPrefix)
	}

	// Sort the filtered tags using semantic versioning
	sortedTags, err := sortTagsByVersion(filteredTags)
	if err != nil {
		return "", fmt.Errorf("failed to sort tags: %v", err)
	}

	return sortedTags[0], nil // Assuming sortedTags is descending
}

// Fetch the latest tag from Bitbucket as a proxy for the latest release
func fetchBitbucketLatestTag(repo string, client *http.Client, baseURL string, token string) (string, error) {
	// Expecting repo in the format 'owner/repo'
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Bitbucket repository format. Expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	if baseURL == "" {
		baseURL = "https://api.bitbucket.org/2.0"
	}
	url := fmt.Sprintf("%s/repositories/%s/%s/refs/tags", baseURL, owner, repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Bitbucket request: %v", err)
	}

	// Optional: Set Authorization header if token is provided via flag or environment variable
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	} else if envToken := os.Getenv("BITBUCKET_TOKEN"); envToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", envToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Bitbucket tags: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting (Bitbucket uses standard HTTP status codes)
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("Bitbucket rate limit exceeded")
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Bitbucket repository not found: %s", repo)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bitbucket API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Bitbucket response body: %v", err)
	}

	var tagsResp BitbucketTagsResponse
	if err := json.Unmarshal(body, &tagsResp); err != nil {
		return "", fmt.Errorf("failed to parse Bitbucket JSON: %v", err)
	}

	if len(tagsResp.Values) == 0 {
		return "", fmt.Errorf("no tags found for Bitbucket repository: %s", repo)
	}

	return tagsResp.Values[0].Name, nil // Assuming the first tag is the latest
}

// Fetch the latest tag from Bitbucket matching a specific version prefix
func fetchBitbucketLatestTagByVersion(repo string, client *http.Client, baseURL string, token string, versionPrefix string) (string, error) {
	// Expecting repo in the format 'owner/repo'
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Bitbucket repository format. Expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	if baseURL == "" {
		baseURL = "https://api.bitbucket.org/2.0"
	}
	perPage := 100
	url := fmt.Sprintf("%s/repositories/%s/%s/refs/tags?pagelen=%d", baseURL, owner, repoName, perPage)

	var allTags []BitbucketTag

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create Bitbucket tags request: %v", err)
		}

		// Set Authorization header
		if token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		} else if envToken := os.Getenv("BITBUCKET_TOKEN"); envToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", envToken))
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch Bitbucket tags: %v", err)
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", fmt.Errorf("Bitbucket rate limit exceeded")
		}

		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("Bitbucket repository not found: %s", repo)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Bitbucket API error: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read Bitbucket tags response body: %v", err)
		}

		var tagsResp BitbucketTagsResponse
		if err := json.Unmarshal(body, &tagsResp); err != nil {
			return "", fmt.Errorf("failed to parse Bitbucket tags JSON: %v", err)
		}

		if len(tagsResp.Values) == 0 {
			break
		}

		allTags = append(allTags, tagsResp.Values...)

		// Check if there is a next page
		if tagsResp.Next == "" {
			break
		}
		url = tagsResp.Next
	}

	// Filter tags by version prefix
	var filteredTags []string
	for _, tag := range allTags {
		if strings.HasPrefix(tag.Name, "php-"+versionPrefix) {
			filteredTags = append(filteredTags, tag.Name)
		}
	}

	if len(filteredTags) == 0 {
		return "", fmt.Errorf("no tags found for version prefix: %s", versionPrefix)
	}

	// Sort the filtered tags using semantic versioning
	sortedTags, err := sortTagsByVersion(filteredTags)
	if err != nil {
		return "", fmt.Errorf("failed to sort tags: %v", err)
	}

	return sortedTags[0], nil // Assuming sortedTags is descending
}

// sortTagsByVersion sorts a slice of tag strings in descending semantic version order
func sortTagsByVersion(tags []string) ([]string, error) {
	type versionTag struct {
		Version *semver.Version
		Tag     string
	}

	var vt []versionTag
	re := regexp.MustCompile(`^php-(v?[\d\.]+.*)$`) // Adjust regex based on tag naming conventions

	for _, tag := range tags {
		matches := re.FindStringSubmatch(tag)
		if len(matches) != 2 {
			continue // Skip tags that don't match the expected pattern
		}
		versionStr := matches[1]
		// Handle tags that start with 'v', e.g., 'v8.2.26'
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			// Skip tags that don't conform to semver
			continue
		}
		vt = append(vt, versionTag{
			Version: version,
			Tag:     tag,
		})
	}

	if len(vt) == 0 {
		return nil, fmt.Errorf("no semver-compatible tags found")
	}

	// Sort the tags in descending order
	sort.Slice(vt, func(i, j int) bool {
		return vt[i].Version.GreaterThan(vt[j].Version)
	})

	var sortedTags []string
	for _, v := range vt {
		sortedTags = append(sortedTags, v.Tag)
	}

	return sortedTags, nil
}
