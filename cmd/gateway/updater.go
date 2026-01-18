package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Updater 鑷洿鏂板櫒
type Updater struct {
	owner string
	repo  string
}

// NewUpdater 鍒涘缓鏇存柊鍣?
func NewUpdater(owner, repo string) *Updater {
	return &Updater{
		owner: owner,
		repo:  repo,
	}
}

// GitHubRelease GitHub Release淇℃伅
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// GetLatestRelease 鑾峰彇鏈€鏂板彂甯冪増鏈?
func (u *Updater) GetLatestRelease() (version string, downloadURL string, err error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("GitHub API杩斿洖閿欒: %d, %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	// 鏍规嵁褰撳墠绯荤粺鏋舵瀯閫夋嫨瀵瑰簲鐨勮祫婧?
	assetName := u.getAssetName()
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return release.TagName, asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("鏈壘鍒伴€傚悎褰撳墠绯荤粺鐨勫彂甯冩枃浠? %s", assetName)
}

// getAssetName 鑾峰彇閫傚悎褰撳墠绯荤粺鐨勮祫婧愬悕绉?
func (u *Updater) getAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("home-gateway-%s-%s%s", goos, goarch, ext)
}

// DownloadAndReplace 涓嬭浇骞舵浛鎹㈠綋鍓嶅彲鎵ц鏂囦欢
func (u *Updater) DownloadAndReplace(downloadURL string) error {
	// 鑾峰彇褰撳墠鍙墽琛屾枃浠惰矾寰?
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("鑾峰彇褰撳墠绋嬪簭璺緞澶辫触: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("瑙ｆ瀽绋嬪簭璺緞澶辫触: %w", err)
	}

	// 涓嬭浇鏂扮増鏈?
	fmt.Println("馃摜 涓嬭浇涓?..")
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("涓嬭浇澶辫触: HTTP %d", resp.StatusCode)
	}

	// 鍒涘缓涓存椂鏂囦欢
	tmpPath := execPath + ".new"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("鍒涘缓涓存椂鏂囦欢澶辫触: %w", err)
	}

	// 涓嬭浇鍒颁复鏃舵枃浠?
	written, err := io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("鍐欏叆涓存椂鏂囦欢澶辫触: %w", err)
	}
	fmt.Printf("   宸蹭笅杞?%d 瀛楄妭\n", written)

	// 璁剧疆鍙墽琛屾潈闄愶紙Linux/macOS锛?
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("璁剧疆鏉冮檺澶辫触: %w", err)
		}
	}

	// 澶囦唤鏃х増鏈?
	backupPath := execPath + ".bak"
	if runtime.GOOS == "windows" {
		// Windows涓嬮渶瑕佸厛鍒犻櫎鏃у浠?
		os.Remove(backupPath)
	}
	
	if err := os.Rename(execPath, backupPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("澶囦唤鏃х増鏈け璐? %w", err)
	}

	// 鏇挎崲涓烘柊鐗堟湰
	if err := os.Rename(tmpPath, execPath); err != nil {
		// 灏濊瘯鎭㈠
		os.Rename(backupPath, execPath)
		return fmt.Errorf("鏇挎崲澶辫触: %w", err)
	}

	// 鍒犻櫎澶囦唤锛堝彲閫夛級
	os.Remove(backupPath)

	return nil
}

// getLatestReleaseForArch 鑾峰彇鎸囧畾鏋舵瀯鐨勪笅杞介摼鎺ワ紙鐢ㄤ簬GitHub Actions锛?
func (u *Updater) getLatestReleaseForArch(goos, goarch string) (downloadURL string, err error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	assetName := fmt.Sprintf("home-gateway-%s-%s%s", goos, goarch, ext)

	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, assetName) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("鏈壘鍒? %s", assetName)
}
