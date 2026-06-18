package src

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubRepo   = "anomalyco/opencode"
	githubAPIURL = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

func UpgradeOpencode() error {
	fmt.Printf("Checking latest opencode version...\n")

	version, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}
	fmt.Printf("Latest version: %s\n", version)

	filename := getDownloadFilename()
	url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", githubRepo, version, filename)

	fmt.Printf("Downloading %s...\n", filename)
	tmpDir, err := os.MkdirTemp("", "opencode-upgrade-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, filename)
	if err := downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	binPath := GetOpencodeBinPath()
	binDir := filepath.Dir(binPath)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	fmt.Printf("Extracting to %s...\n", binPath)
	if err := extractBinary(archivePath, binPath); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	fmt.Printf("Upgrade complete: opencode v%s\n", version)
	return nil
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

func getDownloadFilename() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}

	ext := ".tar.gz"
	if osName == "darwin" || osName == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("opencode-%s-%s%s", osName, arch, ext)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archivePath, binPath string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, binPath)
	}
	return extractTarGz(archivePath, binPath)
}

func extractTarGz(archivePath, binPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if filepath.Base(hdr.Name) == "opencode" && hdr.Typeflag == tar.TypeReg {
			out, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, tr)
			return err
		}
	}
	return fmt.Errorf("opencode binary not found in archive")
}

func extractZip(archivePath, binPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	target := "opencode"
	if runtime.GOOS == "windows" {
		target = "opencode.exe"
	}

	for _, f := range r.File {
		if filepath.Base(f.Name) == target {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("opencode binary not found in archive")
}
