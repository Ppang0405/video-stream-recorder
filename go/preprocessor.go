package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Preprocessor types
type YTDLPProcessor struct {
	binaryPath string
	timeout    time.Duration
	quality    string
}

type VideoInfo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Duration  float64  `json:"duration"`
	Formats   []Format `json:"formats"`
	IsLive    bool     `json:"is_live"`
	Thumbnail string   `json:"thumbnail"`
}

type Format struct {
	FormatID   string  `json:"format_id"`
	URL        string  `json:"url"`
	Protocol   string  `json:"protocol"`
	Extension  string  `json:"ext"`
	Resolution string  `json:"resolution"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Filesize   int64   `json:"filesize"`
	TBR        float64 `json:"tbr"`
}

type StreamInfo struct {
	HLSURL     string
	Title      string
	Duration   float64
	Resolution string
	IsLive     bool
}

// NewYTDLPProcessor creates a new yt-dlp processor
func NewYTDLPProcessor(binaryPath, quality string) *YTDLPProcessor {
	return &YTDLPProcessor{
		binaryPath: binaryPath,
		timeout:    30 * time.Second,
		quality:    quality,
	}
}

// isHLSURL checks if the URL is already an HLS stream
func isHLSURL(urlStr string) bool {
	return strings.Contains(strings.ToLower(urlStr), ".m3u8") ||
		strings.Contains(strings.ToLower(urlStr), "m3u8")
}

// needsPreprocessing determines if URL needs preprocessing
func needsPreprocessing(urlStr string) bool {
	if isHLSURL(urlStr) {
		return false
	}
	
	// Common video platforms that need preprocessing
	platforms := []string{
		"youtube.com", "youtu.be",
		"twitch.tv",
		"vimeo.com",
		"dailymotion.com",
		"facebook.com",
		"instagram.com",
		"tiktok.com",
		"twitter.com", "x.com",
	}
	
	for _, platform := range platforms {
		if strings.Contains(strings.ToLower(urlStr), platform) {
			return true
		}
	}
	
	return false
}

// ExtractVideoInfo extracts video metadata using yt-dlp
func (p *YTDLPProcessor) ExtractVideoInfo(inputURL string) (*VideoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, p.binaryPath,
		"-J",
		"--no-warnings",
		"--no-playlist",
		inputURL)
	
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v, stderr: %s", err, stderr.String())
	}
	
	var info VideoInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %v", err)
	}
	
	return &info, nil
}

// ExtractHLSURL extracts the best HLS stream URL
func (p *YTDLPProcessor) ExtractHLSURL(inputURL string) (*StreamInfo, error) {
	// First check if already HLS
	if isHLSURL(inputURL) {
		return &StreamInfo{
			HLSURL: inputURL,
			Title:  "Direct HLS Stream",
		}, nil
	}
	
	// Extract video info
	info, err := p.ExtractVideoInfo(inputURL)
	if err != nil {
		return nil, err
	}
	
	// Find best HLS format
	var bestFormat *Format
	var bestBitrate float64
	
	for _, format := range info.Formats {
		if format.Protocol == "m3u8" || format.Protocol == "m3u8_native" {
			// Prefer higher bitrate formats
			if bestFormat == nil || format.TBR > bestBitrate {
				bestFormat = &format
				bestBitrate = format.TBR
			}
		}
	}
	
	if bestFormat == nil {
		// Fallback: try to get any HLS URL using yt-dlp format selector
		hlsURL, err := p.getHLSURLDirect(inputURL)
		if err != nil {
			return nil, fmt.Errorf("no HLS stream found in formats")
		}
		return &StreamInfo{
			HLSURL:   hlsURL,
			Title:    info.Title,
			Duration: info.Duration,
			IsLive:   info.IsLive,
		}, nil
	}
	
	return &StreamInfo{
		HLSURL:     bestFormat.URL,
		Title:      info.Title,
		Duration:   info.Duration,
		Resolution: bestFormat.Resolution,
		IsLive:     info.IsLive,
	}, nil
}

// getHLSURLDirect uses yt-dlp format selector to get HLS URL directly
func (p *YTDLPProcessor) getHLSURLDirect(inputURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	
	formatSelector := fmt.Sprintf("%s[protocol=m3u8]", p.quality)
	
	cmd := exec.CommandContext(ctx, p.binaryPath,
		"-f", formatSelector,
		"--get-url",
		"--no-warnings",
		"--no-playlist",
		inputURL)
	
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		// Fallback to any HLS stream
		cmd = exec.CommandContext(ctx, p.binaryPath,
			"-f", "best[protocol=m3u8]",
			"--get-url",
			"--no-warnings",
			"--no-playlist",
			inputURL)
		
		stdout.Reset()
		stderr.Reset()
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("no HLS stream available: %v, stderr: %s", err, stderr.String())
		}
	}
	
	url := strings.TrimSpace(stdout.String())
	if url == "" {
		return "", fmt.Errorf("empty URL returned from yt-dlp")
	}
	
	return url, nil
}

// TestYTDLP tests if yt-dlp is available and working
func TestYTDLP(binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, binaryPath, "--version")
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found or not working: %v", err)
	}
	
	version := strings.TrimSpace(stdout.String())
	log.Printf("yt-dlp version: %s", version)
	
	return nil
}

// ProcessURL handles the preprocessing of input URLs
func ProcessURL(processor *YTDLPProcessor, inputURL string, enablePreprocess bool) (string, error) {
	if !enablePreprocess {
		return inputURL, nil
	}
	
	if !needsPreprocessing(inputURL) {
		log.Printf("URL doesn't need preprocessing: %s", inputURL)
		return inputURL, nil
	}
	
	log.Printf("Preprocessing URL with yt-dlp: %s", inputURL)
	
	streamInfo, err := processor.ExtractHLSURL(inputURL)
	if err != nil {
		return "", fmt.Errorf("preprocessing failed: %v", err)
	}
	
	log.Printf("Extracted HLS URL: %s", streamInfo.HLSURL)
	if streamInfo.Title != "" {
		log.Printf("Video title: %s", streamInfo.Title)
	}
	if streamInfo.IsLive {
		log.Printf("Live stream detected")
	}
	
	return streamInfo.HLSURL, nil
}
