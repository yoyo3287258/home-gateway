package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// Updater 自动更新器
type Updater struct {
	owner string
	repo  string
}

// NewUpdater 创建更新器
func NewUpdater(owner, repo string) *Updater {
	return &Updater{
		owner: owner,
		repo:  repo,
	}
}

// GitHubRelease GitHub Release信息
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// GetLatestRelease 获取最新版本信息
func (u *Updater) GetLatestRelease() (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API返回错误: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	// 寻找匹配当前平台的Asset
	assetName := fmt.Sprintf("home-gateway-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return release.TagName, "", fmt.Errorf("未找到适用于 %s-%s 的发布包", runtime.GOOS, runtime.GOARCH)
	}

	return release.TagName, downloadURL, nil
}

// DownloadAndReplace 下载并替换当前程序
func (u *Updater) DownloadAndReplace(url string) error {
	// 1. 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 2. 准备临时文件
	currentExe, err := os.Executable()
	if err != nil {
		return err
	}

	tmpFile := currentExe + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// 3. 写入内容
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close() // 显式关闭以确保写入完成

	// 4. 替换
	oldFile := currentExe + ".old"
	
	// 如果存在旧的.old文件，先删除
	os.Remove(oldFile)

	// 重命名当前文件为.old
	if err := os.Rename(currentExe, oldFile); err != nil {
		return fmt.Errorf("备份旧文件失败: %w", err)
	}

	// 将新文件重命名为当前文件
	if err := os.Rename(tmpFile, currentExe); err != nil {
		// 回滚
		os.Rename(oldFile, currentExe)
		return fmt.Errorf("替换文件失败: %w", err)
	}

	// 尝试添加可执行权限（Unix/Linux）
	if runtime.GOOS != "windows" {
		os.Chmod(currentExe, 0755)
	}

	return nil
}
