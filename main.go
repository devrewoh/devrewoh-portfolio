package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

// Database connection pool
var dbPool *pgxpool.Pool

// Server represents the HTTP server configuration
type Server struct {
	router chi.Router
	addr   string
	logger *slog.Logger
}

func init() {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// initDB initializes the database connection pool
func initDB() error {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	dbPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	return dbPool.Ping(context.Background())
}

// generateAPIKey creates a new API key and stores it in the database
func generateAPIKey(ctx context.Context, email, tier string, credits int) (string, error) {
	// Generate random key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Create key with prefix
	key := "ic_" + hex.EncodeToString(b)

	// Hash for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	// Store in database
	query := `
		INSERT INTO api_keys (key_hash, key_prefix, user_email, name, monthly_limit)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := dbPool.Exec(ctx, query, keyHash, key[:10], email, tier, credits)
	if err != nil {
		return "", err
	}

	return key, nil
}

// NewServer creates a new server instance with configured routes
func NewServer(addr string) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	s := &Server{
		router: chi.NewRouter(),
		addr:   addr,
		logger: logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Compress(5))
	s.router.Use(middleware.Timeout(30 * time.Second))
	s.router.Use(s.securityMiddleware)
	s.router.Use(middleware.Throttle(100)) // Rate limiting
}

// loggingMiddleware provides structured logging
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// securityMiddleware adds comprehensive security headers
func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"style-src 'self' 'unsafe-inline' fonts.googleapis.com; "+
				"font-src 'self' fonts.gstatic.com; "+
				"img-src 'self' data: https:; "+
				"script-src 'self'")

		// HSTS for HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// setupRoutes configures the application routes
func (s *Server) setupRoutes() {
	// Static files with cache headers
	s.router.Handle("/static/*", s.staticFileHandler())

	// Page routes
	s.router.Get("/", s.handleHome)
	s.router.Get("/about", s.handleAbout)
	s.router.Get("/contact", s.handleContact)
	s.router.Get("/compress", s.handleCompress)
	s.router.Get("/compress/docs", s.handleDocs)
	s.router.Post("/checkout", s.handleCheckout)
	s.router.Get("/compress/success", s.handleSuccess)

	// Health check
	s.router.Get("/health", s.handleHealth)

	// 404 handler
	s.router.NotFound(s.handle404)
}

// staticFileHandler serves static files with appropriate headers
func (s *Server) staticFileHandler() http.Handler {
	fileServer := http.FileServer(http.Dir("./static/"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers for static assets
		if strings.HasSuffix(r.URL.Path, ".css") ||
			strings.HasSuffix(r.URL.Path, ".js") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".jpg") ||
			strings.HasSuffix(r.URL.Path, ".ico") {
			w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours
		}

		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	})
}

// renderTemplate safely renders a template with error handling
func (s *Server) renderTemplate(w http.ResponseWriter, r *http.Request, component templ.Component, pageName string) {
	if err := component.Render(r.Context(), w); err != nil {
		s.logger.Error("template render error",
			"page", pageName,
			"error", err,
			"ip", r.RemoteAddr,
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Page handlers
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	component := HomePage("Chris", "Backend developer building with Go")
	s.renderTemplate(w, r, component, "home")
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	component := AboutPage()
	s.renderTemplate(w, r, component, "about")
}

func (s *Server) handleContact(w http.ResponseWriter, r *http.Request) {
	component := ContactPage()
	s.renderTemplate(w, r, component, "contact")
}

func (s *Server) handleCompress(w http.ResponseWriter, r *http.Request) {
	component := CompressPage()
	s.renderTemplate(w, r, component, "compress")
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	component := DocsPage()
	s.renderTemplate(w, r, component, "docs")
}

func (s *Server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tier from form
	tier := r.FormValue("tier")

	// Map tier to Stripe Price ID
	priceID := ""
	switch tier {
	case "starter":
		priceID = os.Getenv("STRIPE_PRICE_STARTER")
	case "growth":
		priceID = os.Getenv("STRIPE_PRICE_GROWTH")
	case "professional":
		priceID = os.Getenv("STRIPE_PRICE_PRO")
	default:
		http.Error(w, "Invalid tier", http.StatusBadRequest)
		return
	}

	// Create Stripe Checkout Session
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://devrewoh.com/compress/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://devrewoh.com/compress"),
	}

	sess, err := session.New(params)
	if err != nil {
		s.logger.Error("stripe session creation failed", "error", err)
		http.Error(w, "Payment processing error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, sess.URL, http.StatusSeeOther)
}

func (s *Server) handleSuccess(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	// Retrieve Stripe session to get customer email and payment details
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		s.logger.Error("failed to retrieve stripe session", "error", err)
		http.Error(w, "Payment verification failed", http.StatusInternalServerError)
		return
	}

	// Determine tier and credits from amount paid
	tier := "unknown"
	credits := 0
	amountPaid := sess.AmountTotal // in cents

	switch amountPaid {
	case 1000: // $10
		tier = "Starter"
		credits = 1500
	case 3900: // $39
		tier = "Growth"
		credits = 10000
	case 9900: // $99
		tier = "Professional"
		credits = 50000
	}

	// Generate API key
	apiKey, err := generateAPIKey(r.Context(), sess.CustomerDetails.Email, tier, credits)
	if err != nil {
		s.logger.Error("failed to generate api key", "error", err)
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Render success page with API key
	component := SuccessPage(apiKey, tier, credits, sess.CustomerDetails.Email)
	s.renderTemplate(w, r, component, "success")
}

func (s *Server) handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	component := NotFoundPage()
	s.renderTemplate(w, r, component, "404")
}

// handleHealth provides a comprehensive health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	// Simple JSON encoding without external deps
	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","version":"%s","uptime":"%s"}`,
		health["status"], health["timestamp"], health["version"], health["uptime"])
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	server := &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Prevent slowloris attacks
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown setup
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		s.logger.Info("server starting", "addr", s.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	s.logger.Info("server shutting down")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown failed", "error", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("server stopped gracefully")
	return nil
}

var startTime = time.Now()

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	// Initialize database
	if err := initDB(); err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	server := NewServer(addr)
	if err := server.Start(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
