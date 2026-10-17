package main

import (
	"fmt"
	"html/template"
	"net/http"
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
var templates map[string]*template.Template

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

func contactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["contact"]
	data := BasePageData{
		Title: "Contact",
	} // SecondaryNav is nil here, so it safely hides on the contact page

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
}

func main() {
	// Initialize template cache on startup
	initTemplates()

	// Register route handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contact", contactHandler)

	http.HandleFunc("/favicon.ico", faviconHandler)

	fmt.Println("Server running at http://localhost:4000...")
	if err := http.ListenAndServe(":4000", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
