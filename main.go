package main

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var pathPattern = regexp.MustCompile(`^/(?:` +
	// version
	`(\d+(?:\.\d+){2}(?:-[^/]+)?)|` +
	// commit
	`(?:sha-)([^/]+)|` +
	// branchOrTag
	`([^/]+)` +
	`)$`)

const (
	versionIndex = iota + 1
	commitIndex
	branchOrTagIndex
)

var client = &http.Client{
	Transport: &http.Transport{
		TLSHandshakeTimeout:   5 * time.Second,
		IdleConnTimeout:       60 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		MaxIdleConns:          5,
		MaxIdleConnsPerHost:   5,
		MaxConnsPerHost:       5,
	},
}

func handle(w http.ResponseWriter, r *http.Request) {
	match := pathPattern.FindStringSubmatch(r.URL.Path)
	if match == nil {
		http.NotFound(w, r)
		return
	}

	if !(r.Method == http.MethodHead || r.Method == http.MethodGet) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var (
		head           string
		shouldRedirect bool
	)
	switch true {
	case match[versionIndex] != "":
		head = "v" + match[versionIndex]
	case match[commitIndex] != "":
		head = match[commitIndex]
		shouldRedirect = true
	case match[branchOrTagIndex] != "":
		head = match[branchOrTagIndex]
		shouldRedirect = true
	default:
		panic("unreachable")
	}

	schemaURL := "https://raw.githubusercontent.com/linearmouse/linearmouse/" + url.PathEscape(head) + "/Documentation/Configuration.json"

	if shouldRedirect {
		http.Redirect(w, r, schemaURL, http.StatusTemporaryRedirect)
		return
	}

	schema, err := getSchema(schemaURL)
	switch err {
	case nil:

	case errSchemaNotFound:
		http.NotFound(w, r)

	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(schema)))
	w.Write(schema)
}

func main() {
	http.HandleFunc("/", handle)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
