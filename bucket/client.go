package bucket

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ListBucketResult represents the XML response from S3 bucket listing
type ListBucketResult struct {
	XMLName        xml.Name        `xml:"ListBucketResult"`
	Name           string          `xml:"Name"`
	Prefix         string          `xml:"Prefix"`
	MaxKeys        int             `xml:"MaxKeys"`
	Delimiter      string          `xml:"Delimiter"`
	IsTruncated    bool            `xml:"IsTruncated"`
	CommonPrefixes []CommonPrefix  `xml:"CommonPrefixes"`
	Contents       []BucketContent `xml:"Contents"`
}

// CommonPrefix represents a folder/prefix in the bucket
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// BucketContent represents a file in the bucket
type BucketContent struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// Client handles S3 bucket operations
type Client struct {
	BaseURL string
}

// NewClient creates a new bucket client
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// ListFolders lists all folders (common prefixes) under a given prefix
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

// ListFiles lists all files under a given prefix
func (c *Client) ListFiles(prefix string) ([]BucketContent, error) {
	result, err := c.listBucket(prefix)
	if err != nil {
		return nil, err
	}
	return result.Contents, nil
}

// ListFilesByTier lists all files under a prefix that match the given tier
func (c *Client) ListFilesByTier(prefix, tier string) ([]BucketContent, error) {
	files, err := c.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	tierPrefix := "combine-" + tier
	filtered := make([]BucketContent, 0)
	for _, f := range files {
		// Extract filename from the full key
		parts := strings.Split(f.Key, "/")
		filename := parts[len(parts)-1]
		if strings.HasPrefix(filename, tierPrefix) {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

// GetAllDemosByTier fetches all demos for a tier across all combine day folders
func (c *Client) GetAllDemosByTier(combinesPrefix, tier string) ([]BucketContent, error) {
	// First, list all day folders
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

// GetDownloadURL returns the full download URL for a file key
func (c *Client) GetDownloadURL(key string) string {
	return c.BaseURL + key
}

func (c *Client) listBucket(prefix string) (*ListBucketResult, error) {
	// Build the URL with query parameters
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
