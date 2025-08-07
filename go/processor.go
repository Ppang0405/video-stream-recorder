package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/grafov/m3u8"
)

// HLSProcessor handles the core HLS stream processing
type HLSProcessor struct {
	url           string
	cache         *FIFOCache
	postProcessor *PostProcessor
	running       bool
}

// NewHLSProcessor creates a new HLS processor
func NewHLSProcessor(streamURL string, postProcessor *PostProcessor) *HLSProcessor {
	return &HLSProcessor{
		url:           streamURL,
		cache:         NewFIFOCache(100), // Initialize with reasonable cache size
		postProcessor: postProcessor,
		running:       false,
	}
}

// Start begins the HLS stream processing
func (p *HLSProcessor) Start() {
	p.running = true
	go p.fetchLoop()
}

// Stop stops the HLS stream processing
func (p *HLSProcessor) Stop() {
	p.running = false
}

// fetchLoop is the main processing loop for HLS streams
func (p *HLSProcessor) fetchLoop() {
	mainurl, _ := url.Parse(p.url)
	
	for p.running {
	start_at:
		b := p.fetch(mainurl.String())
		if b == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		
		buf := bytes.NewBuffer(b)
		pl, pt, err := m3u8.Decode(*buf, true)
		if err != nil {
			log.Printf("fetcher error: %v %v", mainurl.String(), err)
			time.Sleep(1 * time.Second)
			continue
		}
		
		if pt == m3u8.MASTER {
			masterpl := pl.(*m3u8.MasterPlaylist)
			for _, variant := range masterpl.Variants {
				mainurl, _ = mainurl.Parse(variant.URI)
				log.Printf("Selected variant: %v", mainurl.String())
				goto start_at
			}
		} else if pt == m3u8.MEDIA {
			mediapl := pl.(*m3u8.MediaPlaylist)
			p.processSegments(mediapl, mainurl)
		}
		
		time.Sleep(3 * time.Second)
	}
}

// processSegments handles individual segment processing
func (p *HLSProcessor) processSegments(playlist *m3u8.MediaPlaylist, baseURL *url.URL) {
	for _, segment := range playlist.Segments {
		if segment == nil {
			continue
		}
		
		fetchurl, _ := baseURL.Parse(segment.URI)
		fetchurl.RawQuery = baseURL.RawQuery
		
		if p.cache.Set(fetchurl.String()) {
			log.Printf("Processing segment: %v", fetchurl.String())
			
			currenttime := time.Now().UnixNano()
			item := &DatabaseItem{
				Name: fmt.Sprintf("%v.ts", currenttime),
				Len:  segment.Duration,
				T:    currenttime,
			}
			
			// Download segment
			segmentData := p.fetch(fetchurl.String())
			if segmentData != nil {
				// Save to filesystem
				filePath := "./files/" + item.Name
				err := ioutil.WriteFile(filePath, segmentData, 0755)
				if err != nil {
					log.Printf("error on write file to fs %v", err)
					continue
				}
				
				// Store metadata in database
				database_store(item)
				
				// Post-process segment if enabled
				if p.postProcessor != nil {
					go p.postProcessor.ProcessSegment(filePath, item)
				}
			}
		}
	}
}

// fetch downloads content from a URL
func (p *HLSProcessor) fetch(url string) []byte {
	hc := http.Client{Timeout: 10 * time.Second}

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", "iptv/1.0")

	response, err := hc.Do(request)
	if err != nil {
		log.Printf("fetch error %v %v", url, err)
		return nil
	}
	defer response.Body.Close()
	
	if response.StatusCode/100 != 2 {
		log.Printf("Invalid response code %v %v", url, response.StatusCode)
		return nil
	}
	
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("read error %v", err)
		return nil
	}
	
	return b
}

// GetStats returns processing statistics
func (p *HLSProcessor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":     p.running,
		"url":         p.url,
		"cache_size":  p.cache.Size(),
	}
}
