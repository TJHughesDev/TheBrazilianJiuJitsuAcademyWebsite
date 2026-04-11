package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
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

// ─── State ─────────────────────────────────────────────────────

var (
	instructors     []Instructor
	schedule        []ClassSession
	timetable       []TimetableDay
	testimonials    []Testimonial
	indexTmpl       *template.Template
	timetableTmpl   *template.Template
	membershipsTmpl *template.Template
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

	log.Printf("server → http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// ─── Data & templates ──────────────────────────────────────────

func loadData() error {
	read := func(path string, v any) error {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, v)
	}

	if err := read("data/instructors.json", &instructors); err != nil {
		return err
	}
	if err := read("data/schedule.json", &schedule); err != nil {
		return err
	}
	if err := read("data/timetable.json", &timetable); err != nil {
		return err
	}
	if err := read("data/testimonials.json", &testimonials); err != nil {
		return err
	}

	return nil
}

var partials = []string{
	"templates/layout.html",
	"templates/partials/section-header.html",
	"templates/partials/cta-strip.html",
	"templates/partials/testimonials.html",
}

func parseTemplate(pages ...string) (*template.Template, error) {
	files := append(partials, pages...)
	return template.New("base").Funcs(funcMap).ParseFiles(files...)
}

func loadTemplates() error {
	var err error
	indexTmpl, err = parseTemplate("templates/index.html")
	if err != nil {
		return err
	}
	timetableTmpl, err = parseTemplate("templates/timetable.html")
	if err != nil {
		return err
	}
	membershipsTmpl, err = parseTemplate("templates/memberships.html")
	return err
}

// ─── Handlers ──────────────────────────────────────────────────

func renderHTML(w http.ResponseWriter, t *template.Template, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "internal error", 500)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, indexTmpl, PageData{
		Title:        "The Brazilian Jiu Jitsu Academy",
		Description:  "BJJ training",
		Page:         "home",
		Testimonials: testimonials,
	})
}

func handleTimetable(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, timetableTmpl, PageData{Page: "timetable", Timetable: timetable})
}

func handleMemberships(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, membershipsTmpl, PageData{Page: "memberships"})
}

func handleInstructors(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, instructors)
}

func handleSchedule(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, schedule)
}

func handleTaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, 405, APIResponse{Error: "method not allowed"})
		return
	}

	var req TasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, 400, APIResponse{Error: "invalid body"})
		return
	}

	if req.Name == "" || req.Email == "" {
		jsonResponse(w, 400, APIResponse{Error: "name and email required"})
		return
	}

	go sendTasterEmail(req)

	jsonResponse(w, 200, APIResponse{
		Status:  "success",
		Message: "Thanks! We'll be in touch shortly.",
	})
}

// ─── Email (FIXED) ─────────────────────────────────────────────

func sendTasterEmail(req TasterRequest) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	to := os.Getenv("TO_EMAIL")

	if host == "" || user == "" || pass == "" {
		log.Println("[email] missing SMTP config")
		return
	}
	if port == "" {
		port = "465"
	}
	if to == "" {
		to = user
	}

	subject := fmt.Sprintf("Taster Request — %s", req.Name)

	body := fmt.Sprintf(
		"Name: %s\nEmail: %s\nPhone: %s\nInterest: %s\n\nMessage:\n%s",
		req.Name, req.Email, req.Phone, req.Interest, req.Message,
	)

	msg := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		to, user, subject, body,
	)

	addr := net.JoinHostPort(host, port)

	var client *smtp.Client
	var err error

	if port == "465" {
		// SSL/TLS
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			log.Println("[email] TLS dial error:", err)
			return
		}
		client, err = smtp.NewClient(conn, host)
	} else {
		// STARTTLS
		client, err = smtp.Dial(addr)
		if err != nil {
			log.Println("[email] dial error:", err)
			return
		}
		if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			log.Println("[email] STARTTLS error:", err)
			return
		}
	}

	if err != nil {
		log.Println("[email] client error:", err)
		return
	}
	defer client.Quit()

	client.Hello("localhost")

	if err = client.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
		log.Println("[email] auth error:", err)
		return
	}

	if err = client.Mail(user); err != nil {
		log.Println("[email] MAIL FROM error:", err)
		return
	}
	if err = client.Rcpt(to); err != nil {
		log.Println("[email] RCPT TO error:", err)
		return
	}

	w, err := client.Data()
	if err != nil {
		log.Println("[email] DATA error:", err)
		return
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Println("[email] write error:", err)
		return
	}

	w.Close()

	log.Println("[email] sent successfully")
}

// ─── Helpers ───────────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}