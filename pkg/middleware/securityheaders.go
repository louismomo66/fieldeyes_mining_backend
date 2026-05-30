package middleware

import (
	"net/http"
	"os"
)

// SecurityHeadersMiddleware adds standard HTTP security headers to every response.
// These headers instruct browsers to enforce security policies that protect users
// against clickjacking, MIME sniffing, XSS, and protocol downgrade attacks.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent the page from being embedded in an iframe (clickjacking protection)
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent browsers from MIME-sniffing the content type
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Force HTTPS for 1 year; include subdomains
		// Only set in production — browsers will cache this and block HTTP
		if os.Getenv("APP_ENV") == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Control how much referrer information is sent with requests
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict access to browser features (camera, microphone, geolocation, etc.)
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Content Security Policy — API only returns JSON so this is strict
		// Blocks any attempt to render the API response as a web page with scripts
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Remove the server fingerprint header added by nginx/caddy
		// (nginx sets this — we override it to hide version info)
		w.Header().Set("X-Powered-By", "")

		next.ServeHTTP(w, r)
	})
}
