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

type PageData struct {
	Title        string
	Description  string
	Heading      string
	Content      string
	Items        []string
	SecondaryNav []NavLink
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse both the base layout and the child template together
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare data including optional Secondary Navigation links
	data := PageData{
		Title:       "Overview",
		Description: "Learn how to build modular websites with Go html/template layouts.",
		Heading:     "Dynamic Head & Layout Management",
		Content:     "Go templates make it simple to combine a master layout structure with individual page templates. The head metadata, active navigation state, and main content blocks are evaluated dynamically at runtime.",
		Items: []string{
			"Dynamic head tags for SEO optimization",
			"Conditional secondary navigation rendering",
			"Responsive CSS grid structure with main and aside areas",
		},
		SecondaryNav: []NavLink{
			{Label: "Overview", URL: "#overview"},
			{Label: "Architecture", URL: "#architecture"},
			{Label: "Performance", URL: "#performance"},
		},
	}

	// Execute the base layout template ("layout.html")
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", homeHandler)

	fmt.Println("Server running at http://localhost:4000...")
	if err := http.ListenAndServe(":4000", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
