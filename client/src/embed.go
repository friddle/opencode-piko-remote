package src

import (
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed embed/opencode.gz
var embeddedOpencode embed.FS

func GetOpencodeInstallDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".opencode", "bin")
}

func GetOpencodeBinPath() string {
	binName := "opencode"
	if runtime.GOOS == "windows" {
		binName = "opencode.exe"
	}
	return filepath.Join(GetOpencodeInstallDir(), binName)
}

func EnsureOpencode() (string, error) {
	binPath := GetOpencodeBinPath()

	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	fmt.Printf("Installing embedded opencode to %s...\n", binPath)
	return binPath, installEmbeddedOpencode(binPath)
}

func installEmbeddedOpencode(binPath string) error {
	dir := filepath.Dir(binPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	gzData, err := embeddedOpencode.Open("embed/opencode.gz")
	if err != nil {
		return fmt.Errorf("open embedded data: %w", err)
	}
	defer gzData.Close()

	gr, err := gzip.NewReader(gzData)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	out, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, gr); err != nil {
		os.Remove(binPath)
		return fmt.Errorf("decompress: %w", err)
	}

	fmt.Printf("Installed opencode to %s\n", binPath)
	return nil
}
