package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge-api-client/version"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/ricardow2005/MetriyonAPI/releases/latest"
	updateCheckInterval    = 24 * time.Hour
)

// UpdateInfo describes the latest official GitHub release compared with the running build.
type UpdateInfo struct {
	Checked         bool      `json:"checked"`
	CurrentVersion  string    `json:"currentVersion"`
	LatestVersion   string    `json:"latestVersion"`
	UpdateAvailable bool      `json:"updateAvailable"`
	ReleaseName     string    `json:"releaseName"`
	ReleaseURL      string    `json:"releaseUrl"`
	Notes           string    `json:"notes"`
	PublishedAt     time.Time `json:"publishedAt"`
	CheckedAt       time.Time `json:"checkedAt"`
}

type updateCheckCache struct {
	CheckedAt time.Time `json:"checkedAt"`
}

type githubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

// CheckForUpdates checks the latest public GitHub Release. Automatic checks are throttled to once every 24 hours.
// Pass force=true for a user-requested manual check.
func (a *App) CheckForUpdates(force bool) (UpdateInfo, error) {
	current := strings.TrimPrefix(strings.TrimSpace(version.Version), "v")
	info := UpdateInfo{CurrentVersion: current}

	cachePath, cacheErr := updateCachePath()
	if cacheErr == nil && !force {
		if cache, err := readUpdateCache(cachePath); err == nil && time.Since(cache.CheckedAt) < updateCheckInterval {
			return info, nil
		}
	}

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubLatestReleaseURL, nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "MetriyonAPI/"+current)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := client.Do(req)
	if err != nil {
		return info, fmt.Errorf("não foi possível consultar atualizações: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return info, fmt.Errorf("GitHub Releases respondeu HTTP %d", response.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return info, fmt.Errorf("resposta inválida do GitHub Releases: %w", err)
	}

	latest := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	info.Checked = true
	info.LatestVersion = latest
	info.UpdateAvailable = isNewerVersion(latest, current)
	info.ReleaseName = release.Name
	info.ReleaseURL = release.HTMLURL
	info.Notes = truncateReleaseNotes(release.Body, 5000)
	info.PublishedAt = release.PublishedAt
	info.CheckedAt = time.Now().UTC()

	if cacheErr == nil {
		_ = writeUpdateCache(cachePath, updateCheckCache{CheckedAt: info.CheckedAt})
	}
	return info, nil
}

// OpenExternalURL opens only official MetriyonAPI GitHub release pages in the default browser.
func (a *App) OpenExternalURL(rawURL string) error {
	const allowedPrefix = "https://github.com/ricardow2005/MetriyonAPI/releases"
	if !strings.HasPrefix(strings.TrimSpace(rawURL), allowedPrefix) {
		return fmt.Errorf("URL externa não permitida")
	}
	if a.ctx == nil || !a.eventsEnabled {
		return fmt.Errorf("janela do aplicativo ainda não está pronta")
	}
	wruntime.BrowserOpenURL(a.ctx, rawURL)
	return nil
}

func updateCachePath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "ForgeAPIClient")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check.json"), nil
}

func readUpdateCache(path string) (updateCheckCache, error) {
	var cache updateCheckCache
	data, err := os.ReadFile(path)
	if err != nil {
		return cache, err
	}
	return cache, json.Unmarshal(data, &cache)
}

func writeUpdateCache(path string, cache updateCheckCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func truncateReleaseNotes(notes string, max int) string {
	notes = strings.TrimSpace(notes)
	if len(notes) <= max {
		return notes
	}
	return strings.TrimSpace(notes[:max]) + "\n…"
}

func isNewerVersion(latest, current string) bool {
	latestParts := numericVersionParts(latest)
	currentParts := numericVersionParts(current)
	length := len(latestParts)
	if len(currentParts) > length {
		length = len(currentParts)
	}
	for i := 0; i < length; i++ {
		var left, right int
		if i < len(latestParts) {
			left = latestParts[i]
		}
		if i < len(currentParts) {
			right = currentParts[i]
		}
		if left != right {
			return left > right
		}
	}
	return false
}

func numericVersionParts(value string) []int {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	if index := strings.IndexAny(value, "-+"); index >= 0 {
		value = value[:index]
	}
	parts := strings.Split(value, ".")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil {
			number = 0
		}
		result = append(result, number)
	}
	return result
}
