package middleware

import (
	"net/http"
	"path"
	"strings"
)

// SecurityHeaders applies security headers, adapting for static assets.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		urlPath := strings.ToLower(r.URL.Path)
		ext := strings.ToLower(path.Ext(urlPath))

		// Detect likely static resources by extension
		// isStatic := ext != "" && (strings.Contains(".js.css.png.jpg.jpeg.webp.gif.svg.ico.mp4.mp3.wav.webm.json.map.ttf.woff.woff2", ext))
		isStatic := ext != "" && (strings.Contains(".png.jpg.jpeg.webp.gif.svg.ico.mp4.mp3.wav.webm.ttf.woff.woff2", ext))

		// Always safe headers
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		if r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		if isStatic {
			// Static file headers: caching-friendly, relaxed CSP
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
			h.Del("Pragma")
			h.Del("Expires")
			h.Set("X-Frame-Options", "SAMEORIGIN")
			h.Set("Content-Security-Policy", "default-src 'self'")
			// Skip COOP/CORP to allow asset embedding
		} else {
			// API / HTML responses: strict mode
			h.Set("X-Frame-Options", "DENY")
			h.Set("Content-Security-Policy",
				"default-src 'self'; "+
					"object-src 'none'; "+
					"base-uri 'self'; "+
					"frame-ancestors 'none'; "+
					"form-action 'self'; "+
					"block-all-mixed-content;")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			h.Set("Cross-Origin-Resource-Policy", "same-origin")
			h.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			h.Set("Pragma", "no-cache")
			h.Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// package middleware

// import "net/http"

// // SecurityHeaders applies a recommended set of HTTP security headers.
// // Notes:
// // - CSP here is strict; adjust if your app needs external resources (CDNs, analytics, etc).
// // - HSTS is applied only when the request is over TLS to avoid problems in local dev.
// func SecurityHeaders(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		h := w.Header()

// 		// Prevent MIME sniffing
// 		h.Set("X-Content-Type-Options", "nosniff")

// 		// Prevent clickjacking
// 		h.Set("X-Frame-Options", "SAMEORIGIN")

// 		// Content Security Policy: strict default, no objects, no framing, self-only forms/base.
// 		// Adjust this if you need to allow scripts/styles from external sources.
// 		h.Set("Content-Security-Policy",
// 			"default-src 'self'; "+
// 				"object-src 'none'; "+
// 				"base-uri 'self'; "+
// 				"frame-ancestors 'none'; "+
// 				"form-action 'self'; "+
// 				"block-all-mixed-content;")

// 		// HSTS only on HTTPS; do not set for plain HTTP requests
// 		if r.TLS != nil {
// 			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
// 		}

// 		// Referrer and feature controls
// 		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
// 		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

// 		// Cross-origin policies to reduce data exfiltration surface
// 		// h.Set("Cross-Origin-Opener-Policy", "same-origin")
// 		// h.Set("Cross-Origin-Resource-Policy", "same-origin")

// 		// Caching: for authenticated API responses it's safer to prevent caching
// 		// h.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
// 		h.Set("Cache-Control", "public, max-age=31536000, immutable")
// 		// h.Set("Pragma", "no-cache")
// 		// h.Set("Expires", "0")

// 		next.ServeHTTP(w, r)
// 	})
// }
