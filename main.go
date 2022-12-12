package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/xpy123993/shorten/store"
)

var (
	data              = flag.String("data", "/var/tmp/store.json", "The json file to store shorten links")
	metadataFolder    = flag.String("metadata-folder", "/var/tmp/", "The folder to save the screenshot. Leave empty to disable the archive feature.")
	metadataPath      = flag.String("metadata-path", "/metadata/", "The route path to store metadata files.")
	archiveWorkerNum  = flag.Int("archive-worker", 2, "Number of archive workers")
	addr              = flag.String("addr", "0.0.0.0:8080", "HTTP address")
	allowedSchemes    = flag.String("scheme-allowlist", "http,https,ftp", "The list of URL scheme that can be shortend.")
	updatePath        = flag.String("update-path", "/update", "The path to insert a link")
	generateURLPrefix = flag.String("prefix", "http://127.0.0.1:8080/", "The prefix of generated shorten url.")
)

func getAbsoluteLink(baseUrl string) string {
	if strings.HasPrefix(*generateURLPrefix, "http") {
		return *generateURLPrefix + baseUrl
	}
	return "//" + *generateURLPrefix + baseUrl
}

func createHandler(schemeAllowList map[string]bool, urlStore *store.Store, archiveChan chan archiveTask) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(*updatePath, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "cannot parse request", http.StatusBadRequest)
			return
		}
		if len(r.FormValue("url")) == 0 {
			fmt.Fprintf(w, `<html><head><title>URL Shortener</title></head>
			<body><form method="post">
			  <label>URL: </label><input placeholder="https://example.com" name="url">
			  <input type="hidden" name="source" value="form"/><input type="submit">
			</form></body></html>`)
			return
		}
		targetURL, err := url.Parse(r.FormValue("url"))
		if err != nil {
			http.Error(w, "cannot parse target URL", http.StatusBadRequest)
			return
		}
		if allow, exist := schemeAllowList[targetURL.Scheme]; !allow || !exist {
			http.Error(w, "URL scheme is not supported", http.StatusNotImplemented)
			return
		}
		token, err := urlStore.AddLink(targetURL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if archiveChan != nil {
			if _, err := os.Stat(path.Join(*metadataFolder, token+".png")); os.IsNotExist(err) {
				archiveChan <- archiveTask{targetURL: targetURL.String(), targetBaseName: token}
			}
		}
		if err := urlStore.DumpToDisk(*data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		url := *generateURLPrefix + token
		if r.FormValue("source") == "form" {
			fmt.Fprintf(w, `<html><head><title>URL Shortener</title></head>
			<body><div>%s</div><script>
			navigator.clipboard.writeText("%s");
			document.body.insertAdjacentText('beforeend', 'Has been copied to the clipboard');
			</script></body></html>`, url, url)
		} else {
			fmt.Fprintf(w, "%s\n", url)
		}
	})
	mux.HandleFunc(*metadataPath, func(w http.ResponseWriter, r *http.Request) {
		requestedResource := strings.TrimPrefix(r.RequestURI, *metadataPath)
		token := requestedResource
		ext := strings.ToLower(path.Ext(requestedResource))
		if len(ext) > 0 {
			if ext != ".png" && ext != ".pdf" {
				http.Error(w, "extension is not supported", http.StatusNotImplemented)
				return
			}
			token = strings.TrimSuffix(requestedResource, ext)
		}
		if matched, err := regexp.MatchString("[a-zA-Z0-9_-]{3,8}", token); err != nil || !matched {
			http.Error(w, "cannot parse target URL", http.StatusBadRequest)
			return
		}
		targetURL, err := urlStore.Query(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !strings.HasPrefix(targetURL, "http") {
			http.Error(w, "archive is not supported for non-http pages", http.StatusNotImplemented)
			return
		}
		if len(ext) == 0 {
			fmt.Fprintf(w, `<html><head><title>URL Shortener</title></head>
			<body><div>
			This link is mapped to <a href='%s'>%s</a></div>
			<div>Possible archives: <a href='%s'>PDF</a> <a href='%s'>PNG</a></div>
			
			</body></html>`, targetURL, targetURL, getAbsoluteLink(strings.TrimPrefix(*metadataPath, "/")+token+".pdf"), getAbsoluteLink(strings.TrimPrefix(*metadataPath, "/")+token+".png"))
			return
		}
		targetFile := path.Join(*metadataFolder, requestedResource)
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			archiveChan <- archiveTask{targetURL: targetURL, targetBaseName: token}
		}
		http.ServeFile(w, r, targetFile)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := urlStore.Query(strings.TrimPrefix(r.RequestURI, "/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
	})
	return mux
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixMicro())

	urlStore, err := store.OpenOrCreate(*data)
	if err != nil {
		log.Fatal(err)
	}
	schemeAllowlist := map[string]bool{}
	for _, scheme := range strings.Split(*allowedSchemes, ",") {
		schemeAllowlist[scheme] = true
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
	defer lis.Close()

	var archiveChan chan archiveTask
	if len(*metadataFolder) > 0 {
		archiveChan = make(chan archiveTask)
		defer close(archiveChan)
		for i := 0; i < *archiveWorkerNum; i++ {
			go createArciveWorker(archiveChan)
		}
	}

	log.Printf("Serving at %v", lis.Addr())
	if err := http.Serve(lis, createHandler(schemeAllowlist, urlStore, archiveChan)); err != nil {
		log.Printf("Server exited with error: %v", err)
	}
}
