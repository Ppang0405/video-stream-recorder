package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	flagURL         = flag.String("url", "", "url to fetch")
	flagTail        = flag.Int("tail", 24, "how much hours to keep")
	flagHost        = flag.String("host", "http://localhos:8080", "add host to m3u8")
	flagBindTo      = flag.String("bind-to", ":8080", "bind to ip:port")
	flagPreprocess  = flag.Bool("preprocess", false, "enable preprocessing with yt-dlp")
	flagYTDLPPath   = flag.String("ytdlp-path", "yt-dlp", "path to yt-dlp binary")
	flagQuality     = flag.String("quality", "best", "video quality preference for yt-dlp")

	flagDebug = flag.Bool("debug", false, "")

	flagVersion = flag.Bool("version", false, "show version")
)

func main() {
	flag.Parse()

	if *flagVersion {
		println("v0.2")
		return
	}

	if *flagURL == "" {
		log.Printf("Set url to fetch: ./app --url [url-here]")
		return
	}

	// Initialize preprocessor if enabled
	var preprocessor *YTDLPProcessor
	if *flagPreprocess {
		preprocessor = NewYTDLPProcessor(*flagYTDLPPath, *flagQuality)
		log.Printf("Preprocessor enabled with yt-dlp path: %s", *flagYTDLPPath)
		
		// Test yt-dlp availability
		if err := TestYTDLP(*flagYTDLPPath); err != nil {
			log.Fatalf("yt-dlp test failed: %v", err)
		}
		log.Printf("yt-dlp test successful")
	}

	// Preprocess the input URL
	processedURL, err := ProcessURL(preprocessor, *flagURL, *flagPreprocess)
	if err != nil {
		log.Fatalf("URL preprocessing failed: %v", err)
	}
	
	log.Printf("Final URL to fetch: %s", processedURL)

	// Create files directory
	err = os.MkdirAll("./files", 0755)
	if err != nil {
		log.Fatalf("mkdir fail: %v", err)
	}

	// Initialize database
	database_init()

	// Initialize post-processor (currently disabled by default)
	postProcessorConfig := PostProcessorConfig{
		Enabled:       false, // Enable in future
		CloudUpload:   false,
		VideoStitch:   false,
		CleanupAfter:  24 * time.Hour,
		OutputFormats: []string{"mp4"},
	}
	postProcessor := NewPostProcessor(postProcessorConfig)

	// Initialize HLS processor
	hlsProcessor := NewHLSProcessor(processedURL, postProcessor)

	// Start background workers
	if !*flagDebug {
		hlsProcessor.Start()
	}
	go database_worker()

	// Initialize and start HTTP server
	httpServer := NewHTTPServer(*flagBindTo, *flagHost, hlsProcessor)
	log.Fatal(httpServer.Start())
}
