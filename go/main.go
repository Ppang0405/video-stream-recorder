package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/grafov/m3u8"
)

var (
	flagURL    = flag.String("url", "", "url to fetch")
	flagTail   = flag.Int("tail", 24, "how much hours to keep")
	flagHost   = flag.String("host", "http://localhos:8080", "add host to m3u8")
	flagBindTo = flag.String("bind-to", ":8080", "bind to ip:port")

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

	err := os.MkdirAll("./files", 0755)
	if err != nil {
		log.Fatalf("mkdir fail: %v", err)
	}

	database_init()

	if !*flagDebug {
		go fetcher()
	}

	go database_worker()

	router := mux.NewRouter()

	router.HandleFunc("/live/stream.m3u8", func(w http.ResponseWriter, r *http.Request) {

		utc := r.URL.Query().Get("utc")
		if utc != "" {
			// w.Header().Set("Location", "/")
			vod1(w, r)
			return
		}

		out := "#EXTM3U\n"
		out += "#EXT-X-TARGETDURATION:2\n"
		out += "#EXT-X-VERSION:4\n"

		items := database_last_5()

		if len(items) <= 0 {
			return
		}

		last := items[len(items)-1]

		out += fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", last.ID)

		for _, item := range items {
			out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
			out += fmt.Sprintf("%s\n", item.Name)
		}

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(out)))
		w.Write([]byte(out))

	})

	router.HandleFunc("/live/{ts:.+}", serve_ts_file)

	router.HandleFunc("/start/{start}/{limit}/vod.m3u8", func(w http.ResponseWriter, r *http.Request) {

		varz := mux.Vars(r)

		out := "#EXTM3U\n"
		out += "#EXT-X-PLAYLIST-TYPE:VOD\n"
		out += "#EXT-X-TARGETDURATION:20\n"
		out += "#EXT-X-VERSION:4\n"
		out += "#EXT-X-MEDIA-SEQUENCE:0\n"

		t, err := time.Parse("20060102150405", varz["start"])
		if err != nil {
			perr, ok := err.(*time.ParseError)
			log.Printf("error %v %v", perr, ok)
			return
		}
		start := fmt.Sprintf("%d", t.Unix())
		println(start)

		items := database_get(start, varz["limit"])

		for _, item := range items {
			out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
			out += fmt.Sprintf("%s\n", item.Name)
		}

		out += "#EXT-X-ENDLIST\n"

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(out)))
		w.Write([]byte(out))
	})

	router.HandleFunc("/start/{start}/{limit}/stream.m3u8", vod1)

	router.HandleFunc("/start/{start}/{limit}/{ts:.+}", serve_ts_file)

	// Add root handler for convenience
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>VSR - Video Stream Recorder</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007cba; padding-bottom: 10px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007cba; }
        .endpoint h3 { margin-top: 0; color: #007cba; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
        .live { border-left-color: #28a745; }
        .vod { border-left-color: #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎥 Video Stream Recorder (VSR)</h1>
        <p>Welcome to VSR! Your HLS stream recorder is running.</p>
        
        <div class="endpoint live">
            <h3>📡 Live Stream</h3>
            <p><strong>HLS Manifest:</strong> <code>/live/stream.m3u8</code></p>
            <p><strong>Example:</strong> <code>http://localhost:8080/live/stream.m3u8</code></p>
        </div>
        
        <div class="endpoint vod">
            <h3>🎬 Video On Demand (VOD)</h3>
            <p><strong>VOD Manifest:</strong> <code>/start/{timestamp}/{duration}/vod.m3u8</code></p>
            <p><strong>Stream Manifest:</strong> <code>/start/{timestamp}/{duration}/stream.m3u8</code></p>
            <p><strong>Timestamp Format:</strong> YYYYMMDDHHMMSS (e.g., 20240107120000)</p>
        </div>
        
        <div class="endpoint">
            <h3>🔗 Quick Links</h3>
            <p><a href="/live/stream.m3u8">Live Stream (M3U8)</a></p>
        </div>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Add /live redirect for convenience
	router.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/live/stream.m3u8", http.StatusFound)
	})

	log.Printf("Starting server on %v", *flagBindTo)
	log.Fatal(http.ListenAndServe(*flagBindTo, router))
}

func vod1(w http.ResponseWriter, r *http.Request) {

	varz := mux.Vars(r)

	utc := r.URL.Query().Get("utc")
	if utc != "" {
		varz["start"] = utc
		varz["limit"] = "300"
	}

	out := "#EXTM3U\n"
	out += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	out += "#EXT-X-TARGETDURATION:20\n"
	out += "#EXT-X-VERSION:4\n"
	out += "#EXT-X-MEDIA-SEQUENCE:0\n"

	items := database_get(varz["start"], varz["limit"])

	for _, item := range items {
		out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
		out += fmt.Sprintf("%s\n", item.Name)
	}

	out += "#EXT-X-ENDLIST\n"

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(out)))
	w.Write([]byte(out))
}

func serve_ts_file(w http.ResponseWriter, r *http.Request) {
	varz := mux.Vars(r)
	w.Header().Set("Content-Type", "text/vnd.trolltech.linguist")
	b, err := ioutil.ReadFile(fmt.Sprintf("./files/%s", varz["ts"]))
	if err != nil {
		log.Printf("error %v", err)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Write(b)

}

func fetcher() {
	mainurl, _ := url.Parse(*flagURL)
	for {
	start_at:
		b := fetch(mainurl.String())
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
				log.Printf("%v", mainurl.String())
				goto start_at
			}

		} else if pt == m3u8.MEDIA {
			mediapl := pl.(*m3u8.MediaPlaylist)
			for _, segment := range mediapl.Segments {
				if segment == nil {
					continue
				}
				fetchurl, _ := mainurl.Parse(segment.URI)
				fetchurl.RawQuery = mainurl.RawQuery
				if cache_set(fetchurl.String()) {
					log.Printf("%v", fetchurl.String())
					currenttime := time.Now().UnixNano()
					item := &DatabaseItem{
						Name: fmt.Sprintf("%v.ts", currenttime),
						Len:  segment.Duration,
						T:    currenttime,
					}
					database_store(item)

					b := fetch(fetchurl.String())
					if b != nil {
						err := ioutil.WriteFile("./files/"+item.Name, b, 0755)
						if err != nil {
							log.Printf("error on write file to fs %v", err)
							continue
						}
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func fetch(url string) []byte {
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
		return nil
	}
	return b
}
