package handlers

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "text/template"
  "time"

  "landing/internal/config"
  "landing/internal/database"
  "landing/internal/email"
)

type Handler struct {
  db            *database.DB
  cfg           *config.Config
  email         *email.Sender
  templatesPath string
  logger        *log.Logger
}

func New(db *database.DB, cfg *config.Config) *Handler {
  templatesPath := "templates"
  if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
    templatesPath = "./templates"
  }

  return &Handler{
    db:            db,
    cfg:           cfg,
    email:         email.New(cfg),
    templatesPath: templatesPath,
    logger:        log.New(os.Stdout, "", log.LstdFlags),
  }
}

// loggingMiddleware logs HTTP requests
func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    userAgent := r.UserAgent()

    // Block malicious bots and common attack patterns
    blockedPatterns := []string{
      "python-requests",
      "curl",
      "wget",
      "sqlmap",
      "nikto",
      ".php",
      ".env",
      ".git",
      "wp-admin",
      "xmlrpc",
      "backup",
      "config",
    }

    for _, pattern := range blockedPatterns {
      if strings.Contains(strings.ToLower(r.RequestURI), strings.ToLower(pattern)) {
        w.WriteHeader(http.StatusForbidden)
        fmt.Fprintf(w, "Access Denied")
        h.logger.Printf("BLOCKED attack: %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)
        return
      }
      if strings.Contains(strings.ToLower(userAgent), strings.ToLower(pattern)) {
        w.WriteHeader(http.StatusForbidden)
        fmt.Fprintf(w, "Access Denied")
        h.logger.Printf("BLOCKED bot: %s from %s", userAgent, r.RemoteAddr)
        return
      }
    }

    start := time.Now()
    wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

    next.ServeHTTP(wrapped, r)

    duration := time.Since(start)
    statusColor := getStatusColor(wrapped.statusCode)
    methodColor := getMethodColor(r.Method)

    h.logger.Printf(
      "%s %s %s %s %s %d %s",
      methodColor+r.Method+"\033[0m",
      r.RequestURI,
      statusColor+fmt.Sprintf("%d", wrapped.statusCode)+"\033[0m",
      duration.String(),
      r.RemoteAddr,
      wrapped.contentLength,
      userAgent,
    )
  })
}

type responseWriter struct {
  http.ResponseWriter
  statusCode    int
  contentLength int
}

func (rw *responseWriter) WriteHeader(code int) {
  rw.statusCode = code
  rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
  rw.contentLength = len(b)
  return rw.ResponseWriter.Write(b)
}

// Color codes for terminal output
func getStatusColor(statusCode int) string {
  switch {
  case statusCode >= 200 && statusCode < 300:
    return "\033[32m" // Green
  case statusCode >= 300 && statusCode < 400:
    return "\033[36m" // Cyan
  case statusCode >= 400 && statusCode < 500:
    return "\033[33m" // Yellow
  case statusCode >= 500:
    return "\033[31m" // Red
  default:
    return "\033[37m" // White
  }
}

func getMethodColor(method string) string {
  switch method {
  case "GET":
    return "\033[34m" // Blue
  case "POST":
    return "\033[32m" // Green
  case "PUT":
    return "\033[33m" // Yellow
  case "DELETE":
    return "\033[31m" // Red
  default:
    return "\033[37m" // White
  }
}

func (h *Handler) Start(host, port string) error {
  mux := http.NewServeMux()

  // Serve static files
  mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

  mux.HandleFunc("/", h.indexHandler)
  mux.HandleFunc("/subscribe", h.subscribeHandler)
  mux.HandleFunc("/unsubscribe", h.unsubscribeHandler)
  mux.HandleFunc("/newsletters", h.newslettersHandler)
  mux.HandleFunc("/newsletter/", h.newsletterDetailHandler)
  mux.HandleFunc("/contact", h.contactHandler)
  mux.HandleFunc("/about", h.aboutHandler)

  // Wrap with logging middleware
  handler := h.loggingMiddleware(mux)

  h.logger.Printf("\033[36m▶ Starting server on http://%s:%s\033[0m", host, port)

  return http.ListenAndServe(host+":"+port, handler)
}

func (h *Handler) getTemplatePath(name string) string {
  return filepath.Join(h.templatesPath, name)
}

func (h *Handler) indexHandler(
  w http.ResponseWriter,
  r *http.Request,
) {
  if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  tmpl, err := template.ParseFiles(
    h.getTemplatePath("base.html"),
    h.getTemplatePath("index.html"),
  )
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
    return
  }

  data := map[string]interface{}{
    "IsHome": true,
  }

  tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *Handler) subscribeHandler(
  w http.ResponseWriter,
  r *http.Request,
) {
  if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  var req struct {
    Email string `json:"email"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "Invalid request", http.StatusBadRequest)
    return
  }

  if req.Email == "" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Email is required",
    })
    return
  }

  if err := h.db.AddSubscriber(r.Context(), req.Email); err != nil {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Email already exists",
    })
    return
  }

  unsubscribeLink := fmt.Sprintf(
    "%s/unsubscribe?email=%s",
    getBaseURL(r),
    req.Email,
  )

  if err := h.email.SendConfirmationEmail(
    req.Email,
    unsubscribeLink,
  ); err != nil {
    h.logger.Printf("❌ Failed to send confirmation email to %s: %v", req.Email, err)
  } else {
    h.logger.Printf("✓ Confirmation email sent to %s", req.Email)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(map[string]string{
    "message": "Email has been added",
  })
}

func (h *Handler) unsubscribeHandler(
  w http.ResponseWriter,
  r *http.Request,
) {
  email := r.URL.Query().Get("email")
  if email == "" {
    http.Error(w, "No email specified", http.StatusBadRequest)
    return
  }

  if err := h.db.RemoveSubscriber(r.Context(), email); err != nil {
    http.Error(
      w,
      fmt.Sprintf(
        "Email %s was not found or already unsubscribed",
        email,
      ),
      http.StatusBadRequest,
    )
    return
  }

  h.logger.Printf("✓ Unsubscribed %s", email)

  w.Header().Set("Content-Type", "text/plain")
  fmt.Fprintf(w, "The email %s has been unsubscribed.", email)
}

func (h *Handler) newslettersHandler(
  w http.ResponseWriter,
  r *http.Request,
) {
  newsletters, err := h.db.GetNewsletters(r.Context())
  if err != nil {
    http.Error(w, "Failed to fetch newsletters", 500)
    return
  }

  tmpl, err := template.ParseFiles(
    h.getTemplatePath("base.html"),
    h.getTemplatePath("newsletters.html"),
  )
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
    return
  }

  tmpl.ExecuteTemplate(w, "base.html", newsletters)
}

func (h *Handler) newsletterDetailHandler(
  w http.ResponseWriter,
  r *http.Request,
) {
  idStr := r.URL.Path[len("/newsletter/"):]
  id, err := strconv.Atoi(idStr)
  if err != nil {
    http.Error(w, "Invalid newsletter ID", http.StatusBadRequest)
    return
  }

  newsletter, err := h.db.GetNewsletter(r.Context(), id)
  if err != nil {
    http.Error(w, "Newsletter not found", http.StatusNotFound)
    return
  }

  tmpl, err := template.ParseFiles(
    h.getTemplatePath("base.html"),
    h.getTemplatePath("newsletter_detail.html"),
  )
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
    return
  }

  tmpl.ExecuteTemplate(w, "base.html", newsletter)
}

func (h *Handler) contactHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet && r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  if r.Method == http.MethodGet {
    tmpl, err := template.ParseFiles(
      h.getTemplatePath("base.html"),
      h.getTemplatePath("contact.html"),
    )
    if err != nil {
      http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
      return
    }

    data := map[string]interface{}{
      "IsContact": true,
    }

    tmpl.ExecuteTemplate(w, "base.html", data)
    return
  }

  if r.Method == http.MethodPost {
    h.handleContactSubmission(w, r)
  }
}

func (h *Handler) handleContactSubmission(w http.ResponseWriter, r *http.Request) {
  // Parse form data
  if err := r.ParseForm(); err != nil {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Failed to parse form data",
    })
    return
  }

  // Extract form fields
  name := strings.TrimSpace(r.FormValue("name"))
  email := strings.TrimSpace(r.FormValue("email"))
  subject := strings.TrimSpace(r.FormValue("subject"))
  message := strings.TrimSpace(r.FormValue("message"))
  subscribe := r.FormValue("subscribe") == "on"

  // Validate required fields
  if name == "" || email == "" || subject == "" || message == "" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "All fields are required",
    })
    return
  }

  // Validate email format (basic validation)
  if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Invalid email address",
    })
    return
  }

  // Validate message length (prevent spam)
  if len(message) < 10 {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Message must be at least 10 characters",
    })
    return
  }

  if len(message) > 5000 {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "Message must be less than 5000 characters",
    })
    return
  }

  // If subscribe checkbox is checked, add to subscribers
  if subscribe {
    if err := h.db.AddSubscriber(r.Context(), email); err != nil {
      // Log but don't fail the contact submission
      h.logger.Printf(
        "ℹ Subscriber %s already exists or failed to add: %v",
        email,
        err,
      )
    } else {
      h.logger.Printf("✓ New subscriber added: %s", email)
    }
  }

  // Send confirmation email to the user
  if err := h.email.SendContactConfirmation(email, name); err != nil {
    h.logger.Printf(
      "❌ Failed to send contact confirmation to %s: %v",
      email,
      err,
    )
  } else {
    h.logger.Printf("✓ Contact confirmation email sent to %s", email)
  }

  // Send notification email to admin
  adminEmail := h.cfg.AdminEmail
  if adminEmail != "" {
    if err := h.email.SendContactNotification(
      adminEmail,
      name,
      email,
      subject,
      message,
    ); err != nil {
      h.logger.Printf(
        "❌ Failed to send contact notification to admin: %v",
        err,
      )
    } else {
      h.logger.Printf("✓ Contact notification sent to admin: %s", adminEmail)
    }
  }

  // Save contact message to database (optional - add to your DB interface)
  if err := h.db.AddContactMessage(r.Context(), name, email, subject, message); err != nil {
    h.logger.Printf(
      "⚠ Failed to save contact message: %v",
      err,
    )
  }

  h.logger.Printf("✓ Contact form submitted by %s (%s)", name, email)

  // Return success response
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(map[string]string{
    "message": "Thank you for your message. We'll get back to you soon!",
  })
}

func (h *Handler) aboutHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  tmpl, err := template.ParseFiles(
    h.getTemplatePath("base.html"),
    h.getTemplatePath("about.html"),
  )
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
    h.logger.Printf("❌ Template parse error: %v", err)
    return
  }

  data := map[string]interface{}{
    "IsAbout": true,
  }

  w.Header().Set("Content-Type", "text/html")
  tmpl.ExecuteTemplate(w, "base.html", data)
}

func getBaseURL(r *http.Request) string {
  scheme := "http"
  if r.TLS != nil {
    scheme = "https"
  }
  return fmt.Sprintf("%s://%s", scheme, r.Host)
}