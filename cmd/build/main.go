package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ─── Data types ────────────────────────────────────────────────

type Instructor struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Belt    string `json:"belt"`
	Bio     string `json:"bio"`
	Image   string `json:"image,omitempty"`
	Lineage string `json:"lineage,omitempty"`
}

type ClassSession struct {
	Day   string `json:"day"`
	Time  string `json:"time"`
	Class string `json:"class"`
	Type  string `json:"type"`
	Gi    bool   `json:"gi"`
}

type TimetableSession struct {
	Time string `json:"time"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TimetableDay struct {
	Day      string             `json:"day"`
	Sessions []TimetableSession `json:"sessions"`
}

type Testimonial struct {
	Name   string `json:"name"`
	Quote  string `json:"quote"`
	Rating int    `json:"rating"`
	Image  string `json:"image,omitempty"`
}

type PageData struct {
	Title        string
	Description  string
	Page         string
	Timetable    []TimetableDay
	Testimonials []Testimonial
}

// ─── Template helpers ──────────────────────────────────────────

var funcMap = template.FuncMap{
	"stars": func(n int) string {
		return strings.Repeat("★", n)
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict: odd number of arguments")
		}
		m := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict: keys must be strings")
			}
			m[key] = values[i+1]
		}
		return m, nil
	},
}

// ─── Data loading ──────────────────────────────────────────────

var partials = []string{
	"templates/layout.html",
	"templates/partials/section-header.html",
	"templates/partials/cta-strip.html",
	"templates/partials/testimonials.html",
}

func loadJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

func parseTemplate(page string) (*template.Template, error) {
	files := append(partials, page)
	return template.New("base").Funcs(funcMap).ParseFiles(files...)
}

// ─── Rendering ─────────────────────────────────────────────────

func renderPage(outPath string, tmplFile string, data PageData) {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		log.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
	}

	tmpl, err := parseTemplate(tmplFile)
	if err != nil {
		log.Fatalf("parse %s: %v", tmplFile, err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create %s: %v", outPath, err)
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, "layout", data); err != nil {
		log.Fatalf("execute %s: %v", outPath, err)
	}

	log.Printf("  wrote %s", outPath)
}

// ─── Asset copying ─────────────────────────────────────────────

func copyFile(src, dst string) {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		log.Fatalf("mkdir for %s: %v", dst, err)
	}
	in, err := os.Open(src)
	if err != nil {
		log.Fatalf("open %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		log.Fatalf("create %s: %v", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		log.Fatalf("copy %s→%s: %v", src, dst, err)
	}
}

func copyDir(src, dst string) {
	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		copyFile(path, target)
		return nil
	})
}

// ─── Entry point ───────────────────────────────────────────────

func main() {
	var (
		timetable    []TimetableDay
		testimonials []Testimonial
	)

	if err := loadJSON("data/timetable.json", &timetable); err != nil {
		log.Fatalf("load timetable: %v", err)
	}
	if err := loadJSON("data/testimonials.json", &testimonials); err != nil {
		log.Fatalf("load testimonials: %v", err)
	}

	if err := os.RemoveAll("dist"); err != nil {
		log.Fatalf("clean dist: %v", err)
	}
	if err := os.MkdirAll("dist", 0755); err != nil {
		log.Fatalf("create dist: %v", err)
	}

	log.Println("rendering pages…")

	renderPage("dist/index.html", "templates/index.html", PageData{
		Title:        "The Brazilian Jiu Jitsu Academy",
		Description:  "BJJ classes for adults and children in Market Drayton, Shropshire. Professional black belt instruction for all ages and abilities. Book a free taster session today.",
		Page:         "home",
		Testimonials: testimonials,
	})

	renderPage("dist/timetable/index.html", "templates/timetable.html", PageData{
		Title:       "Class Timetable — The Brazilian Jiu Jitsu Academy",
		Description: "Full weekly class timetable for The Brazilian Jiu Jitsu Academy in Market Drayton. Adult Gi, No-Gi, beginners, and kids sessions.",
		Page:        "timetable",
		Timetable:   timetable,
	})

	renderPage("dist/memberships/index.html", "templates/memberships.html", PageData{
		Title:        "Membership Options — The Brazilian Jiu Jitsu Academy",
		Description:  "Flexible BJJ memberships for all levels at The Brazilian Jiu Jitsu Academy. Monthly plans, drop-in sessions, children's memberships, and armed forces rates.",
		Page:         "memberships",
		Testimonials: testimonials,
	})

	log.Println("copying assets…")
	copyDir("static", "dist/static")

	log.Println("copying data…")
	os.MkdirAll("dist/data", 0755)
	for _, f := range []string{"instructors.json", "schedule.json", "timetable.json", "testimonials.json"} {
		copyFile("data/"+f, "dist/data/"+f)
	}

	log.Println("done → dist/")
}
