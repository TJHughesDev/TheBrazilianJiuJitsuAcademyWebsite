package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// ─── Data types ────────────────────────────────────────────────

type Instructor struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Belt    string `json:"belt"`
	Bio     string `json:"bio"`
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
	Type string `json:"type"` // open-mat | kids | beginners | adults-gi | adults-nogi | competition
}

type TimetableDay struct {
	Day      string             `json:"day"`
	Sessions []TimetableSession `json:"sessions"`
}

type TasterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Interest string `json:"interest"`
	Message  string `json:"message"`
}

type APIResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PageData is passed to every template execution.
type PageData struct {
	Title       string
	Description string
	Page        string // "home" | "timetable" | "memberships"
	Timetable   []TimetableDay
}

// ─── State ─────────────────────────────────────────────────────

var (
	instructors      []Instructor
	schedule         []ClassSession
	timetable        []TimetableDay
	indexTmpl        *template.Template
	timetableTmpl    *template.Template
	membershipsTmpl  *template.Template
)

// ─── Bootstrap ─────────────────────────────────────────────────

func main() {
	if err := loadData(); err != nil {
		log.Fatal(err)
	}
	if err := loadTemplates(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/timetable", handleTimetable)
	mux.HandleFunc("/memberships", handleMemberships)
	mux.HandleFunc("/api/instructors", handleInstructors)
	mux.HandleFunc("/api/schedule", handleSchedule)
	mux.HandleFunc("/api/taster", handleTaster)
	mux.HandleFunc("/", handleIndex)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("BJJA server → http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func loadData() error {
	raw, err := os.ReadFile("data/instructors.json")
	if err != nil {
		return fmt.Errorf("reading instructors.json: %w", err)
	}
	if err := json.Unmarshal(raw, &instructors); err != nil {
		return fmt.Errorf("parsing instructors.json: %w", err)
	}

	raw, err = os.ReadFile("data/schedule.json")
	if err != nil {
		return fmt.Errorf("reading schedule.json: %w", err)
	}
	if err := json.Unmarshal(raw, &schedule); err != nil {
		return fmt.Errorf("parsing schedule.json: %w", err)
	}

	raw, err = os.ReadFile("data/timetable.json")
	if err != nil {
		return fmt.Errorf("reading timetable.json: %w", err)
	}
	if err := json.Unmarshal(raw, &timetable); err != nil {
		return fmt.Errorf("parsing timetable.json: %w", err)
	}

	log.Printf("loaded %d instructors, %d schedule slots, %d timetable days",
		len(instructors), len(schedule), len(timetable))
	return nil
}

func loadTemplates() error {
	var err error

	indexTmpl, err = template.ParseFiles(
		"templates/layout.html",
		"templates/partials/testimonials.html",
		"templates/index.html",
	)
	if err != nil {
		return fmt.Errorf("parsing index templates: %w", err)
	}

	timetableTmpl, err = template.ParseFiles(
		"templates/layout.html",
		"templates/partials/testimonials.html",
		"templates/timetable.html",
	)
	if err != nil {
		return fmt.Errorf("parsing timetable templates: %w", err)
	}

	membershipsTmpl, err = template.ParseFiles(
		"templates/layout.html",
		"templates/partials/testimonials.html",
		"templates/memberships.html",
	)
	if err != nil {
		return fmt.Errorf("parsing memberships templates: %w", err)
	}

	return nil
}

// ─── Handlers ──────────────────────────────────────────────────

func renderHTML(w http.ResponseWriter, t *template.Template, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Println("template error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderHTML(w, indexTmpl, PageData{
		Title:       "The Brazilian Jiu Jitsu Academy | Market Drayton, Shropshire",
		Description: "Professional Brazilian Jiu Jitsu training in Market Drayton, Shropshire. Classes for adults and children from age 6. Book your free taster session today.",
		Page:        "home",
	})
}

func handleTimetable(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, timetableTmpl, PageData{
		Title:       "Class Timetable | The Brazilian Jiu Jitsu Academy",
		Description: "Full weekly class timetable for The Brazilian Jiu Jitsu Academy, Market Drayton. Adult gi, no-gi, kids classes and open mat sessions.",
		Page:        "timetable",
		Timetable:   timetable,
	})
}

func handleMemberships(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, membershipsTmpl, PageData{
		Title:       "Memberships | The Brazilian Jiu Jitsu Academy",
		Description: "Flexible BJJ membership options for adults and children at The Brazilian Jiu Jitsu Academy, Market Drayton. Armed forces discounts available.",
		Page:        "memberships",
	})
}

func handleInstructors(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, instructors)
}

func handleSchedule(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, schedule)
}

func handleTaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	var req TasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, APIResponse{Error: "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.Name == "" || req.Email == "" {
		jsonResponse(w, http.StatusBadRequest, APIResponse{Error: "name and email are required"})
		return
	}

	go sendTasterEmail(req)

	jsonResponse(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Thank you! We'll be in touch shortly to confirm your session.",
	})
}

// ─── Helpers ───────────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sendTasterEmail(req TasterRequest) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	to := os.Getenv("TO_EMAIL")

	if to == "" {
		to = "info@thebrazilianjiujitsuacademy.com"
	}
	if host == "" {
		log.Printf("[email] SMTP_HOST not set — skipping for %s <%s>", req.Name, req.Email)
		return
	}
	if port == "" {
		port = "587"
	}

	subject := fmt.Sprintf("Free Taster Request — %s", req.Name)
	body := fmt.Sprintf("New free taster session request received %s\n\nName:     %s\nEmail:    %s\nPhone:    %s\nInterest: %s\n\nMessage:\n%s\n\n---\nSent via The Brazilian Jiu Jitsu Academy website\n",
		time.Now().Format("02 Jan 2006 at 15:04 MST"),
		req.Name, req.Email, req.Phone, req.Interest, req.Message)

	msg := fmt.Sprintf(
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nTo: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		to, user, subject, body)

	addr := net.JoinHostPort(host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	if err := smtp.SendMail(addr, auth, user, []string{to}, []byte(msg)); err != nil {
		log.Printf("[email] send failed for %s <%s>: %v", req.Name, req.Email, err)
		return
	}
	log.Printf("[email] sent for %s <%s>", req.Name, req.Email)
}
