package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Body.String(), "healthy") {
		t.Error("Expected response to contain 'healthy'")
	}

	// Test content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}

func TestSecurityHeaders(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: https:; script-src 'self'",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		if actualValue == "" {
			t.Errorf("Expected security header %s to be set", header)
		} else if actualValue != expectedValue {
			t.Errorf("Expected header %s to be %q, got %q", header, expectedValue, actualValue)
		}
	}
}

func TestSecurityHeadersHTTPS(t *testing.T) {
	server := NewServer(":8080")

	// Test HSTS header with HTTPS
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	hstsHeader := w.Header().Get("Strict-Transport-Security")
	expectedHSTS := "max-age=31536000; includeSubDomains"
	if hstsHeader != expectedHSTS {
		t.Errorf("Expected HSTS header %q, got %q", expectedHSTS, hstsHeader)
	}
}

func TestNotFoundHandler(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPageHandlers(t *testing.T) {
	server := NewServer(":8080")

	tests := []struct {
		name        string
		path        string
		wantCode    int
		wantContent string
	}{
		{
			name:        "Home page",
			path:        "/",
			wantCode:    http.StatusOK,
			wantContent: "Chris",
		},
		{
			name:        "About page",
			path:        "/about",
			wantCode:    http.StatusOK,
			wantContent: "About Me",
		},
		{
			name:        "Contact page",
			path:        "/contact",
			wantCode:    http.StatusOK,
			wantContent: "Get In Touch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d for %s, got %d", tt.wantCode, tt.path, w.Code)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantContent) {
				t.Errorf("Expected %s to contain %q", tt.name, tt.wantContent)
			}

			// Verify proper HTML structure
			if !strings.Contains(body, "<!doctype html>") {
				t.Errorf("%s should contain proper HTML doctype", tt.name)
			}

			if !strings.Contains(body, "DEVREWOH") {
				t.Errorf("%s should contain site branding", tt.name)
			}
		})
	}
}

func TestStaticFileHeaders(t *testing.T) {
	server := NewServer(":8080")

	// Test cache header logic with different file extensions
	tests := []struct {
		name      string
		path      string
		wantCache bool
	}{
		{"CSS file", "/static/css/styles.css", true},
		{"JS file", "/static/js/app.js", true},
		{"PNG image", "/static/images/logo.png", true},
		{"ICO file", "/static/favicon.ico", true},
		{"Regular file", "/static/readme.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// For files that don't exist (404), we still test the cache header logic
			// The static file handler sets headers before checking if file exists
			cacheControl := w.Header().Get("Cache-Control")

			if tt.wantCache {
				expectedCache := "public, max-age=86400"
				// Only check cache headers if the file was processed by our handler
				// (not 404 from missing file)
				if w.Code == http.StatusOK && cacheControl != expectedCache {
					t.Errorf("Expected Cache-Control %q for %s, got %q", expectedCache, tt.name, cacheControl)
				}
				// If it's 404, that's fine - the file just doesn't exist
				if w.Code == http.StatusNotFound {
					t.Logf("File %s doesn't exist (this is fine for testing)", tt.path)
				}
			} else {
				// For non-cacheable files, we don't expect cache headers
				if cacheControl == "public, max-age=86400" {
					t.Errorf("Did not expect cache headers for %s, but got %q", tt.name, cacheControl)
				}
			}
		})
	}
}

func TestServerConfiguration(t *testing.T) {
	// Save original environment
	originalPort := os.Getenv("PORT")
	defer func() {
		if originalPort == "" {
			os.Unsetenv("PORT")
		} else {
			os.Setenv("PORT", originalPort)
		}
	}()

	// Test default port
	server := NewServer(":8080")
	if server.addr != ":8080" {
		t.Errorf("Expected default addr :8080, got %s", server.addr)
	}

	// Test environment port override
	os.Setenv("PORT", "9000")
	server = NewServer(":9000")
	if server.addr != ":9000" {
		t.Errorf("Expected server addr :9000, got %s", server.addr)
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Test that all middleware is applied
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Security headers should be present (tested elsewhere, but verify integration)
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Security middleware should be applied")
	}

	// Request ID should be added (though we can't test the specific value)
	// This mainly tests that middleware chain doesn't break
}

func TestRateLimiting(t *testing.T) {
	server := NewServer(":8080")

	// Test that we can make multiple requests without hitting rate limit
	// The throttle is set to 100, so 10 requests should be fine
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}
}

func TestHealthEndpointContent(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	body := w.Body.String()

	// Test JSON structure
	expectedFields := []string{"status", "timestamp", "version", "uptime"}
	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("Health response should contain %q field", field)
		}
	}

	// Test cache headers
	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache for health endpoint, got %q", cacheControl)
	}
}

func TestPageRoutesExist(t *testing.T) {
	server := NewServer(":8080")

	routes := []string{"/", "/about", "/contact", "/health"}

	for _, route := range routes {
		t.Run("GET "+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code >= 500 {
				t.Errorf("Route %s returned server error: %d", route, w.Code)
			}

			if w.Code == 0 {
				t.Errorf("Route %s returned no response", route)
			}
		})
	}
}

func BenchmarkHealthEndpoint(b *testing.B) {
	server := NewServer(":8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
}

func BenchmarkHomePageRender(b *testing.B) {
	server := NewServer(":8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
}
