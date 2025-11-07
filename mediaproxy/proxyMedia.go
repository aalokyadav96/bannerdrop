package mediaproxy

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	cacheDir      = "./cache/media"  // directory to store cached files
	cacheMaxAge   = 24 * time.Hour   // duration before cache considered stale
	clientTimeout = 10 * time.Second // fetch timeout
)

func ProxyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	raw := strings.TrimPrefix(ps.ByName("url"), "/")

	var target string
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		target = raw
	} else {
		target = strings.Replace(raw, "/", "://", 1)
	}

	u, err := url.Parse(target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	host := u.Hostname()
	if host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "172.") {
		http.Error(w, "blocked host", http.StatusForbidden)
		return
	}

	// ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		http.Error(w, "cache dir error", http.StatusInternalServerError)
		return
	}

	// create a deterministic filename from the URL hash
	cachePath := filepath.Join(cacheDir, hashURL(u.String()))

	// if cached file exists and fresh, serve it directly
	if fi, err := os.Stat(cachePath); err == nil {
		if time.Since(fi.ModTime()) < cacheMaxAge {
			f, err := os.Open(cachePath)
			if err == nil {
				defer f.Close()
				http.ServeContent(w, r, "", fi.ModTime(), f)
				return
			}
		}
	}

	// fetch from network
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		http.Error(w, "failed request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "MediaProxy/1.0 (+https://yourapp.local)")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{Timeout: clientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		http.Error(w, "remote error", http.StatusBadGateway)
		return
	}

	// save response body to cache file
	tmpPath := cachePath + ".tmp"
	out, err := os.Create(tmpPath)
	if err == nil {
		_, copyErr := io.Copy(out, resp.Body)
		out.Close()
		if copyErr == nil {
			os.Rename(tmpPath, cachePath) // atomic replace
		} else {
			os.Remove(tmpPath)
		}
	} else {
		io.Copy(io.Discard, resp.Body)
	}

	// serve response to client
	for k, v := range resp.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(resp.StatusCode)
	f, _ := os.Open(cachePath)
	if f != nil {
		defer f.Close()
		io.Copy(w, f)
	}
}

// hashURL creates a short filename-safe hash of the URL
func hashURL(u string) string {
	h := sha1.New()
	h.Write([]byte(u))
	return hex.EncodeToString(h.Sum(nil))
}

// package mediaproxy

// import (
// 	"io"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"

// 	"github.com/julienschmidt/httprouter"
// )

// func ProxyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	raw := strings.TrimPrefix(ps.ByName("url"), "/")

// 	var target string
// 	// Handle both /https/i.imgur.com/... and /https://i.imgur.com/... forms
// 	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
// 		target = raw
// 	} else {
// 		target = strings.Replace(raw, "/", "://", 1)
// 	}

// 	u, err := url.Parse(target)
// 	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
// 		http.Error(w, "invalid url", http.StatusBadRequest)
// 		return
// 	}

// 	// Reject private or loopback IPs to prevent SSRF
// 	host := u.Hostname()
// 	if host == "localhost" ||
// 		host == "127.0.0.1" ||
// 		host == "::1" ||
// 		strings.HasPrefix(host, "10.") ||
// 		strings.HasPrefix(host, "192.168.") ||
// 		strings.HasPrefix(host, "172.") {
// 		http.Error(w, "blocked host", http.StatusForbidden)
// 		return
// 	}

// 	// Create a new request with safe headers
// 	req, err := http.NewRequest("GET", u.String(), nil)
// 	if err != nil {
// 		http.Error(w, "failed request", http.StatusInternalServerError)
// 		return
// 	}
// 	req.Header.Set("User-Agent", "MediaProxy/1.0 (+https://yourapp.local)")
// 	req.Header.Set("Accept", "*/*")

// 	// Use custom client with timeout
// 	client := &http.Client{
// 		Timeout: 10 * time.Second,
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		http.Error(w, "fetch failed", http.StatusBadGateway)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// Copy relevant headers
// 	if ctype := resp.Header.Get("Content-Type"); ctype != "" {
// 		w.Header().Set("Content-Type", ctype)
// 	}
// 	if clen := resp.Header.Get("Content-Length"); clen != "" {
// 		w.Header().Set("Content-Length", clen)
// 	}
// 	if cc := resp.Header.Get("Cache-Control"); cc != "" {
// 		w.Header().Set("Cache-Control", cc)
// 	} else {
// 		w.Header().Set("Cache-Control", "public, max-age=86400")
// 	}

// 	w.WriteHeader(resp.StatusCode)
// 	io.Copy(w, resp.Body)
// }
