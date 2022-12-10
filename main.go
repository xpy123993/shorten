package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xpy123993/shorten/store"
)

var (
	data           = flag.String("data", "/var/tmp/store.json", "The json file to store shorten links")
	addr           = flag.String("addr", "0.0.0.0:8080", "HTTP address")
	allowedSchemes = flag.String("scheme-allowlist", "http,https,ftp", "The list of URL scheme that can be shortend.")

	generateURLPrefix = flag.String("prefix", "http://127.0.0.1:8080/", "The prefix of generated shorten url.")
)

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
	http.HandleFunc("/api/v1/add", func(w http.ResponseWriter, r *http.Request) {
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
		if allow, exist := schemeAllowlist[targetURL.Scheme]; !allow || !exist {
			http.Error(w, "URL scheme is not supported", http.StatusNotImplemented)
			return
		}
		token, err := urlStore.AddLink(targetURL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := urlStore.Query(strings.TrimPrefix(r.RequestURI, "/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
	})
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
	log.Printf("Serving at %v", lis.Addr())
	if err := http.Serve(lis, nil); err != nil {
		log.Printf("Server exited with error: %v", err)
	}
}
