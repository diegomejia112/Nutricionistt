package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

const (
	githubRepo = "diegomejia11/nutricionist"
	checkEvery = 6 * time.Hour
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckAndUpdate compares current version with latest GitHub release.
// If newer, downloads and applies the update then signals restart via the returned channel.
func CheckAndUpdate(currentVersion string) (newVersion string, available bool, err error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	var rel ghRelease
	if err = json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", false, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")
	if !isNewer(latest, current) {
		return latest, false, nil
	}

	// Find the right asset for this OS/arch
	assetName := assetNameFor()
	var downloadURL string
	for _, a := range rel.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return latest, true, fmt.Errorf("asset %s no encontrado en release", assetName)
	}

	// Download and apply
	dlResp, err := client.Get(downloadURL)
	if err != nil {
		return latest, true, err
	}
	defer dlResp.Body.Close()

	if err = selfupdate.Apply(dlResp.Body, selfupdate.Options{}); err != nil {
		return latest, true, err
	}

	return latest, true, nil
}

// StartBackgroundChecker runs CheckAndUpdate periodically. onUpdate is called
// with the new version string when an update is successfully applied.
func StartBackgroundChecker(currentVersion string, onUpdate func(newVer string)) {
	go func() {
		// First check after 10 seconds to not block startup
		time.Sleep(10 * time.Second)
		for {
			newVer, available, err := CheckAndUpdate(currentVersion)
			if err == nil && available {
				onUpdate(newVer)
				return // process will restart
			}
			_ = err
			_ = newVer
			time.Sleep(checkEvery)
		}
	}()
}

func assetNameFor() string {
	os_ := runtime.GOOS
	arch := runtime.GOARCH
	ext := ""
	if os_ == "windows" {
		ext = ".exe"
	}
	// e.g. nutricionist-windows-386.exe
	return fmt.Sprintf("nutricionist-%s-%s%s", os_, arch, ext)
}

// isNewer returns true if a > b using simple semver comparison.
func isNewer(a, b string) bool {
	partsA := versionParts(a)
	partsB := versionParts(b)
	for i := 0; i < 3; i++ {
		if partsA[i] > partsB[i] {
			return true
		}
		if partsA[i] < partsB[i] {
			return false
		}
	}
	return false
}

func versionParts(v string) [3]int {
	var parts [3]int
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}

// RestartSelf re-executes the current binary (called after successful update).
func RestartSelf() {
	exec, _ := os.Executable()
	// On Windows, spawn a new process and exit
	_ = exec
	os.Exit(0) // The OS will not restart it — caller should use a launcher bat
}
