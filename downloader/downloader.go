package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Downloader handles downloading and extracting demo files
type Downloader struct {
	OutputDir string
}

// NewDownloader creates a new downloader
func NewDownloader(outputDir string) *Downloader {
	return &Downloader{OutputDir: outputDir}
}

// DownloadResult contains information about a downloaded demo
type DownloadResult struct {
	URL      string
	ZipPath  string
	DemoPath string
	Error    error
}

// Download downloads a file from the given URL and saves it to the output directory
func (d *Downloader) Download(url string) (*DownloadResult, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(d.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract filename from URL
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	zipPath := filepath.Join(d.OutputDir, filename)

	// Check if already downloaded
	if info, err := os.Stat(zipPath); err == nil {
		log.Printf("    Zip already exists: %s (%.2f MB)", filename, float64(info.Size())/(1024*1024))
		return &DownloadResult{
			URL:     url,
			ZipPath: zipPath,
		}, nil
	}

	log.Printf("    Downloading: %s", filename)

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	// Log expected size if available
	if resp.ContentLength > 0 {
		log.Printf("    Size: %.2f MB", float64(resp.ContentLength)/(1024*1024))
	}

	// Create the output file
	out, err := os.Create(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", zipPath, err)
	}
	defer out.Close()

	// Copy with progress tracking
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", zipPath, err)
	}

	log.Printf("    Downloaded: %.2f MB", float64(written)/(1024*1024))

	return &DownloadResult{
		URL:     url,
		ZipPath: zipPath,
	}, nil
}

// Extract extracts a zip file and returns the path to the .dem file
func (d *Downloader) Extract(zipPath string) (string, error) {
	// Open the zip file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer r.Close()

	log.Printf("    Zip contains %d files", len(r.File))

	// Find and extract the .dem file
	for _, f := range r.File {
		log.Printf("    Found in zip: %s (%.2f MB)", f.Name, float64(f.UncompressedSize64)/(1024*1024))

		if strings.HasSuffix(f.Name, ".dem") {
			// Extract directly to output directory (no subdirectory)
			demoPath := filepath.Join(d.OutputDir, filepath.Base(f.Name))

			// Check if already extracted
			if info, err := os.Stat(demoPath); err == nil {
				log.Printf("    Demo already extracted: %s (%.2f MB)", filepath.Base(demoPath), float64(info.Size())/(1024*1024))
				return demoPath, nil
			}

			log.Printf("    Extracting: %s", filepath.Base(f.Name))

			// Extract the file
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file in zip: %w", err)
			}

			out, err := os.Create(demoPath)
			if err != nil {
				rc.Close()
				return "", fmt.Errorf("failed to create demo file: %w", err)
			}

			written, err := io.Copy(out, rc)
			rc.Close()
			out.Close()

			if err != nil {
				return "", fmt.Errorf("failed to extract demo file: %w", err)
			}

			log.Printf("    Extracted: %s (%.2f MB)", filepath.Base(demoPath), float64(written)/(1024*1024))
			return demoPath, nil
		}
	}

	return "", fmt.Errorf("no .dem file found in %s", zipPath)
}

// DownloadAndExtract downloads a zip file and extracts the demo
func (d *Downloader) DownloadAndExtract(url string) (string, error) {
	result, err := d.Download(url)
	if err != nil {
		return "", err
	}

	demoPath, err := d.Extract(result.ZipPath)
	if err != nil {
		return "", err
	}

	return demoPath, nil
}

// Cleanup removes downloaded zip files (keeps extracted demos)
func (d *Downloader) Cleanup() error {
	entries, err := os.ReadDir(d.OutputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
			zipPath := filepath.Join(d.OutputDir, entry.Name())
			if err := os.Remove(zipPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", zipPath, err)
			}
		}
	}

	return nil
}
