package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestContactFormValidation(t *testing.T) {
	server := NewServer(":8080")

	tests := []struct {
		name     string
		formData url.Values
		wantCode int
	}{
		{
			name: "valid form",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"john@example.com"},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusSeeOther, // Redirect on success
		},
		{
			name: "missing name",
			formData: url.Values{
				"email":   {"john@example.com"},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty name",
			formData: url.Values{
				"name":    {""},
				"email":   {"john@example.com"},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"invalid-email"},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty email",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {""},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing message",
			formData: url.Values{
				"name":  {"John Doe"},
				"email": {"john@example.com"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty message",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"john@example.com"},
				"message": {""},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "message too short",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"john@example.com"},
				"message": {"short"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "name too long",
			formData: url.Values{
				"name":    {strings.Repeat("a", 101)}, // 101 characters
				"email":   {"john@example.com"},
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "email too long",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {strings.Repeat("a", 250) + "@example.com"}, // > 254 characters
				"message": {"This is a test message that is long enough to pass validation"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "message too long",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"john@example.com"},
				"message": {strings.Repeat("a", 1001)}, // 1001 characters
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/contact", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d for test case: %s", tt.wantCode, w.Code, tt.name)
			}
		})
	}
}

func TestValidateContactForm(t *testing.T) {
	server := NewServer(":8080")

	tests := []struct {
		name       string
		data       ContactFormData
		wantError  bool
		errorCount int // Expected number of validation errors
	}{
		{
			name: "valid data",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "john@example.com",
				Message: "This is a valid message that meets the minimum length requirement",
			},
			wantError:  false,
			errorCount: 0,
		},
		{
			name: "empty name",
			data: ContactFormData{
				Name:    "",
				Email:   "john@example.com",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "name with whitespace only",
			data: ContactFormData{
				Name:    "   ",
				Email:   "john@example.com",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "invalid email",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "not-an-email",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "empty email",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "message too short",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "john@example.com",
				Message: "short",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "empty message",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "john@example.com",
				Message: "",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "multiple validation errors",
			data: ContactFormData{
				Name:    "",
				Email:   "invalid-email",
				Message: "short",
			},
			wantError:  true,
			errorCount: 3,
		},
		{
			name: "name too long",
			data: ContactFormData{
				Name:    strings.Repeat("a", 101),
				Email:   "john@example.com",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "email too long",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   strings.Repeat("a", 250) + "@example.com",
				Message: "Valid message here that is long enough",
			},
			wantError:  true,
			errorCount: 1,
		},
		{
			name: "message too long",
			data: ContactFormData{
				Name:    "John Doe",
				Email:   "john@example.com",
				Message: strings.Repeat("a", 1001),
			},
			wantError:  true,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := server.validateContactForm(tt.data)
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Expected error: %v, got error: %v, errors: %v",
					tt.wantError, hasError, errors)
			}

			if len(errors) != tt.errorCount {
				t.Errorf("Expected %d errors, got %d errors: %v",
					tt.errorCount, len(errors), errors)
			}
		})
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

func TestPageRoutesExist(t *testing.T) {
	server := NewServer(":8080")

	routes := []string{"/health"}

	for _, route := range routes {
		t.Run("GET "+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code >= 500 {
				t.Errorf("Route %s returned server error: %d", route, w.Code)
			}
		})
	}
}

func TestContactFormTooBig(t *testing.T) {
	server := NewServer(":8080")

	// Create a form that exceeds the 32KB limit
	largeData := strings.Repeat("a", 33*1024) // 33KB
	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"message": {largeData},
	}

	req := httptest.NewRequest("POST", "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for oversized form, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEmailRegex(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.com", true},
		{"user123@example-domain.com", true},
		{"invalid.email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
		{"test..test@example.com", false}, // consecutive dots
		{"test@example", false},           // no TLD
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			match := emailRegex.MatchString(tt.email)
			if match != tt.valid {
				t.Errorf("Email %q: expected valid=%v, got valid=%v", tt.email, tt.valid, match)
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

func BenchmarkContactFormValidation(b *testing.B) {
	server := NewServer(":8080")
	data := ContactFormData{
		Name:    "John Doe",
		Email:   "john@example.com",
		Message: "This is a test message that meets all validation requirements",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.validateContactForm(data)
	}
}

func BenchmarkEmailRegex(b *testing.B) {
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emailRegex.MatchString(email)
	}
}
