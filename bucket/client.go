// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package bucket provides a client for interacting with cloud storage buckets
// (specifically DigitalOcean Spaces) to list and download CS2 demo files.
// It handles XML parsing of bucket listings and filtering demos by competitive tier.
package bucket

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ListBucketResult represents the XML response from an S3-compatible bucket listing.
// It contains metadata about the bucket and its contents.
type ListBucketResult struct {
	XMLName        xml.Name        `xml:"ListBucketResult"`
	Name           string          `xml:"Name"`           // Bucket name
	Prefix         string          `xml:"Prefix"`         // Prefix filter used in the request
	MaxKeys        int             `xml:"MaxKeys"`        // Maximum number of keys returned
	Delimiter      string          `xml:"Delimiter"`      // Delimiter used for grouping (usually "/")
	IsTruncated    bool            `xml:"IsTruncated"`    // Whether results are truncated
	CommonPrefixes []CommonPrefix  `xml:"CommonPrefixes"` // Virtual "folders" in the bucket
	Contents       []BucketContent `xml:"Contents"`       // Actual files in the bucket
}

// CommonPrefix represents a virtual folder/directory in the bucket listing.
type CommonPrefix struct {
	Prefix string `xml:"Prefix"` // The folder path prefix
}

// BucketContent represents a single file/object in the bucket.
type BucketContent struct {
	Key          string `xml:"Key"`          // Full path/key of the object
	LastModified string `xml:"LastModified"` // ISO 8601 timestamp of last modification
	ETag         string `xml:"ETag"`         // Entity tag (hash) for the object
	Size         int64  `xml:"Size"`         // Size in bytes
	StorageClass string `xml:"StorageClass"` // Storage class (e.g., STANDARD)
}

// Client provides methods for interacting with an S3-compatible bucket.
type Client struct {
	BaseURL string // Base URL of the bucket (e.g., https://bucket.nyc3.digitaloceanspaces.com/)
}

// NewClient creates a new bucket client with the specified base URL.
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// ListFolders returns all virtual folders (common prefixes) under the given prefix.
// This is used to enumerate combine event folders in the bucket.
func (c *Client) ListFolders(prefix string) ([]string, error) {
	result, err := c.listBucket(prefix)
	if err != nil {
		return nil, err
	}

	folders := make([]string, 0, len(result.CommonPrefixes))
	for _, cp := range result.CommonPrefixes {
		folders = append(folders, cp.Prefix)
	}
	return folders, nil
}

// ListFiles returns all files under the given prefix.
func (c *Client) ListFiles(prefix string) ([]BucketContent, error) {
	result, err := c.listBucket(prefix)
	if err != nil {
		return nil, err
	}
	return result.Contents, nil
}

// ListFilesByTier returns files filtered by competitive tier.
// It looks for files with names starting with "combine-{tier}".
func (c *Client) ListFilesByTier(prefix, tier string) ([]BucketContent, error) {
	files, err := c.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	tierPrefix := "combine-" + tier
	filtered := make([]BucketContent, 0)
	for _, f := range files {
		parts := strings.Split(f.Key, "/")
		filename := parts[len(parts)-1]
		if strings.HasPrefix(filename, tierPrefix) {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

// GetAllDemosByTier retrieves all demo files for a specific tier across all combine folders.
// It iterates through each combine folder and collects matching demos.
func (c *Client) GetAllDemosByTier(combinesPrefix, tier string) ([]BucketContent, error) {
	folders, err := c.ListFolders(combinesPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list combine folders: %w", err)
	}

	var allDemos []BucketContent
	for _, folder := range folders {
		demos, err := c.ListFilesByTier(folder, tier)
		if err != nil {
			return nil, fmt.Errorf("failed to list files in %s: %w", folder, err)
		}
		allDemos = append(allDemos, demos...)
	}

	return allDemos, nil
}

// GetDownloadURL constructs the full download URL for a given object key.
func (c *Client) GetDownloadURL(key string) string {
	return c.BaseURL + key
}

// listBucket performs the actual HTTP request to list bucket contents.
// It uses the delimiter "/" to enable folder-like navigation.
func (c *Client) listBucket(prefix string) (*ListBucketResult, error) {
	params := url.Values{}
	params.Set("delimiter", "/")
	params.Set("prefix", prefix)

	reqURL := c.BaseURL + "?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bucket listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result ListBucketResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	return &result, nil
}
