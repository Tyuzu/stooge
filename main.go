package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

type NavLink struct {
	Label string
	URL   string
}

type FeatureCard struct {
	Icon        string
	Title       string
	Description string
	ButtonURL   string
	ButtonText  string
}

// BasePageData contains properties available across pages
type BasePageData struct {
	Title        string
	SecondaryNav []NavLink
}

// HomePageData extends base data with specific card features
type HomePageData struct {
	BasePageData
	Heading    string
	FeatureOne FeatureCard
	FeatureTwo FeatureCard
}

// Global template cache map
var (
	templates map[string]*template.Template
	db        *sql.DB
)

// 1. Panic Recovery Middleware
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic and stack trace for debugging
				log.Printf("[PANIC RECOVERED] %v\n%s", err, debug.Stack())

				// Return a safe HTTP 500 Internal Server Error to the client
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Optional: Logging Middleware for debugging request paths and durations
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s | Duration: %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func initTemplates() {
	templates = make(map[string]*template.Template)

	// Custom template function 'dict' for passing multi-argument maps to partials
	funcs := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call: odd number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	pages := []string{"index", "about", "contact"}

	for _, page := range pages {
		// Parse layout, specific page template, and component partials together
		tmpl, err := template.New("layout.html").Funcs(funcs).ParseFiles(
			"templates/layout.html",
			fmt.Sprintf("templates/%s.html", page),
			"templates/components/button.html",
			"templates/components/card.html",
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse template %s: %v", page, err))
		}
		templates[page] = tmpl
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["index"]
	data := HomePageData{
		BasePageData: BasePageData{
			Title: "Home",
			SecondaryNav: []NavLink{
				{Label: "Overview", URL: "/"},
				{Label: "Architecture", URL: "/about"},
			},
		},
		Heading: "Modular Component Architecture",
		FeatureOne: FeatureCard{
			Icon:        "⚡",
			Title:       "High Performance",
			Description: "Compiled straight to machine code with zero external framework overhead.",
			ButtonURL:   "/about",
			ButtonText:  "Learn More",
		},
		FeatureTwo: FeatureCard{
			Icon:        "🛡️",
			Title:       "Context-Aware Escaping",
			Description: "Protects automatically against XSS across HTML, JavaScript, and CSS contexts.",
			ButtonURL:   "/contact",
			ButtonText:  "Get Support",
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["about"]
	data := BasePageData{
		Title: "About Us",
		SecondaryNav: []NavLink{
			{Label: "Team", URL: "/about"},
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ContactForm struct {
	Name    string
	Email   string
	Message string
}

type ContactPageData struct {
	BasePageData
	Success bool
	Error   string
	Form    ContactForm
}

func initDatabase() {
	var err error
	// Open or create a local SQLite database file named "messages.db"
	db, err = sql.Open("sqlite", "./messages.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Verify database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Create the contacts table if it doesn't already exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Connected to SQLite database successfully!")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["contact"]

	if r.Method == http.MethodGet {
		data := ContactPageData{BasePageData: BasePageData{Title: "Contact Us"}}
		_ = tmpl.ExecuteTemplate(w, "layout.html", data)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		form := ContactForm{
			Name:    r.FormValue("name"),
			Email:   r.FormValue("email"),
			Message: r.FormValue("message"),
		}

		// Validate fields
		if form.Name == "" || form.Email == "" || form.Message == "" {
			data := ContactPageData{
				BasePageData: BasePageData{Title: "Contact Us"},
				Error:        "All fields are required. Please complete the form.",
				Form:         form,
			}
			_ = tmpl.ExecuteTemplate(w, "layout.html", data)
			return
		}

		// Save submission to SQLite database using parameterized queries (prevents SQL injection)
		query := `INSERT INTO contacts (name, email, message) VALUES (?, ?, ?)`
		_, err := db.Exec(query, form.Name, form.Email, form.Message)
		if err != nil {
			data := ContactPageData{
				BasePageData: BasePageData{Title: "Contact Us"},
				Error:        "Failed to save message to database. Please try again.",
				Form:         form,
			}
			_ = tmpl.ExecuteTemplate(w, "layout.html", data)
			return
		}

		// Render success state
		data := ContactPageData{
			BasePageData: BasePageData{Title: "Contact Us"},
			Success:      true,
		}
		_ = tmpl.ExecuteTemplate(w, "layout.html", data)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// Handled or ignored gracefully
}

func main() {
	// Initialize database and template cache
	initDatabase()
	initTemplates()

	// Create a new ServeMux router
	mux := http.NewServeMux()

	// Register static file server handler
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register route handlers
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/contact", contactHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)

	// Wrap our router with the Panic Recovery Middleware (and optional Logging Middleware)
	handler := RecoveryMiddleware(LoggingMiddleware(mux))

	fmt.Println("Server running at http://localhost:4000...")
	if err := http.ListenAndServe(":4000", handler); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
