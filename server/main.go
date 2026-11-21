// ap-study: AP科目B 学習ナビ. A single-user, server-rendered (htmx) study
// tracker backed by SQLite. Mounted under BASE_PATH (default /ap) behind Caddy.
package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

type App struct {
	store    *Store
	base     string
	password string
	secret   []byte
	secure   bool
	tmpl     *template.Template
	assetV   string // content hash appended to static URLs for cache busting
	google   *GoogleConfig
	jst      *time.Location
	clock    func() time.Time

	mu       sync.Mutex
	failures []time.Time
}

const cookieName = "ap_session"
const sessionTTL = 90 * 24 * time.Hour

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	if len(os.Args) > 1 {
		os.Exit(runCLI(os.Args[1:]))
	}
	password, secret := os.Getenv("AP_PASSWORD"), os.Getenv("AP_SECRET")
	google := &GoogleConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}
	for _, e := range strings.Split(os.Getenv("ALLOWED_EMAILS"), ",") {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			google.Allowed = append(google.Allowed, e)
		}
	}
	switch {
	case len(secret) < 32:
		slog.Error("AP_SECRET (>=32 chars) is required")
		os.Exit(1)
	case google.enabled() && (google.ClientSecret == "" || google.RedirectURL == "" || len(google.Allowed) == 0):
		slog.Error("GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL and ALLOWED_EMAILS are required with GOOGLE_CLIENT_ID")
		os.Exit(1)
	case !google.enabled() && len(password) < 8:
		slog.Error("set GOOGLE_CLIENT_ID (Google sign-in) or AP_PASSWORD (>=8 chars, local passphrase)")
		os.Exit(1)
	}
	if google.enabled() {
		password = "" // passphrase login is disabled when Google sign-in is configured
	}
	store, err := openStore(cfg("DB_PATH", "data/ap-study.db"))
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	app, err := newApp(store, strings.TrimRight(cfg("BASE_PATH", "/ap"), "/"), password, secret, cfg("COOKIE_SECURE", "1") == "1")
	if err != nil {
		slog.Error("init", "err", err)
		os.Exit(1)
	}
	app.google = google
	slog.Info("auth", "google", google.enabled(), "passphrase", password != "")

	srv := &http.Server{
		Addr:              cfg("ADDR", ":8080"),
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr, "base", app.base)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()
	select {
	case err := <-serveErr:
		slog.Error("serve", "err", err)
		_ = store.db.Close()
		os.Exit(1)
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	_ = store.db.Close()
}

func cfg(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func newApp(store *Store, base, password, secret string, secure bool) (*App, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("JST", 9*3600)
	}
	a := &App{store: store, base: base, password: password, secret: []byte(secret), secure: secure, jst: jst, clock: time.Now, google: &GoogleConfig{}}
	h := sha256.New()
	for _, f := range []string{"static/app.css", "static/app.js", "static/htmx.min.js"} {
		b, err := staticFS.ReadFile(f)
		if err != nil {
			return nil, err
		}
		h.Write(b)
	}
	a.assetV = hex.EncodeToString(h.Sum(nil))[:10]
	a.tmpl, err = template.New("").Funcs(a.funcs()).ParseFS(templateFS, "templates/*.html")
	return a, err
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	b := a.base
	static, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET "+b+"/static/", http.StripPrefix(b+"/static/", cacheStatic(http.FileServerFS(static))))
	mux.HandleFunc("GET "+b+"/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("GET "+b+"/login", a.loginPage)
	mux.HandleFunc("POST "+b+"/login", a.login)
	mux.HandleFunc("GET "+b+"/auth/google", a.googleStart)
	mux.HandleFunc("GET "+b+"/auth/callback", a.googleCallback)
	mux.HandleFunc("GET "+b, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, b+"/", http.StatusMovedPermanently) })

	auth := http.NewServeMux()
	auth.HandleFunc("GET "+b+"/{$}", a.today)
	auth.HandleFunc("GET "+b+"/day/{date}", a.dayPage)
	auth.HandleFunc("GET "+b+"/plan", a.planPage)
	auth.HandleFunc("GET "+b+"/log", a.logPage)
	auth.HandleFunc("GET "+b+"/fields", a.fieldsPage)
	auth.HandleFunc("GET "+b+"/phrases", a.phrasesPage)
	auth.HandleFunc("GET "+b+"/exam", a.examPage)
	auth.HandleFunc("POST "+b+"/logout", a.logout)
	auth.HandleFunc("POST "+b+"/tasks/{id}/toggle", a.toggleTask)
	auth.HandleFunc("POST "+b+"/custom", a.addCustom)
	auth.HandleFunc("POST "+b+"/custom/{id}/toggle", a.toggleCustom)
	auth.HandleFunc("POST "+b+"/custom/{id}/delete", a.deleteCustom)
	auth.HandleFunc("POST "+b+"/sessions", a.addSession)
	auth.HandleFunc("POST "+b+"/sessions/{id}/delete", a.deleteSession)
	auth.HandleFunc("POST "+b+"/fields", a.saveFields)
	auth.HandleFunc("POST "+b+"/mocks/{id}", a.saveMock)
	mux.Handle(b+"/", a.requireAuth(auth))

	return securityHeaders(accessLog(mux))
}

// ---- middleware ----

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
		// same-origin (not no-referrer): no-referrer makes browsers send
		// "Origin: null" on form posts, which the CSRF check rejects.
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		slog.Info("http_request", "method", r.Method, "path", r.URL.Path, "status", sw.status, "duration_ms", time.Since(start).Milliseconds())
	})
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (a *App) authed(r *http.Request) bool {
	_, ok := a.sessionEmail(r)
	return ok
}

// sameOrigin rejects cross-site form posts (CSRF). Browsers send Origin on
// POST; Sec-Fetch-Site is a second signal.
func sameOrigin(r *http.Request) bool {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		return false
	}
	if o := r.Header.Get("Origin"); o != "" {
		u, err := url.Parse(o)
		if err != nil || u.Host != r.Host {
			return false
		}
	}
	return true
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authed(r) {
			if r.Method == http.MethodGet {
				http.Redirect(w, r, a.base+"/login", http.StatusSeeOther)
				return
			}
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", a.base+"/login")
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost && !sameOrigin(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// ---- auth handlers ----

func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	if a.authed(r) {
		http.Redirect(w, r, a.base+"/", http.StatusSeeOther)
		return
	}
	msg := map[string]string{
		"1":       "パスフレーズが違います。",
		"denied":  "このGoogleアカウントでは利用できません。",
		"cancel":  "ログインがキャンセルされました。",
		"expired": "ログインの有効期限が切れました。もう一度お試しください。",
		"failed":  "ログインに失敗しました。もう一度お試しください。",
	}[r.URL.Query().Get("e")]
	a.render(w, "login", map[string]any{"Error": msg, "Google": a.google.enabled(), "Passphrase": a.password != ""})
}

func (a *App) tooManyFailures() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	cut := time.Now().Add(-15 * time.Minute)
	kept := a.failures[:0]
	for _, f := range a.failures {
		if f.After(cut) {
			kept = append(kept, f)
		}
	}
	a.failures = kept
	return len(a.failures) >= 10
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if a.password == "" {
		http.NotFound(w, r)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	if a.tooManyFailures() {
		http.Error(w, "しばらく時間をおいてから再度お試しください。", http.StatusTooManyRequests)
		return
	}
	pw := r.PostFormValue("password")
	if subtle.ConstantTimeCompare([]byte(pw), []byte(a.password)) != 1 {
		a.mu.Lock()
		a.failures = append(a.failures, time.Now())
		a.mu.Unlock()
		time.Sleep(time.Second)
		slog.Warn("login_failed")
		http.Redirect(w, r, a.base+"/login?e=1", http.StatusSeeOther)
		return
	}
	a.setSession(w, passphraseIdentity)
	http.Redirect(w, r, a.base+"/", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: a.base + "/", MaxAge: -1, HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, a.base+"/login", http.StatusSeeOther)
}

// ---- helpers ----

func (a *App) todayStr() string { return a.clock().In(a.jst).Format("2006-01-02") }

func (a *App) render(w http.ResponseWriter, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Base"] = a.base
	data["V"] = a.assetV
	data["Countdown"] = a.countdown()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tmpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("render", "template", name, "err", err)
	}
}

func (a *App) countdown() string {
	today, _ := time.ParseInLocation("2006-01-02", a.todayStr(), a.jst)
	exam, _ := time.ParseInLocation("2006-01-02", days[len(days)-1].Date, a.jst)
	n := int(exam.Sub(today).Hours() / 24)
	switch {
	case n > 0:
		return "本番まであと" + strconv.Itoa(n) + "日"
	case n == 0:
		return "本日本番"
	default:
		return "試験終了"
	}
}

func (a *App) fail(w http.ResponseWriter, err error) {
	slog.Error("handler", "err", err)
	http.Error(w, "サーバーでエラーが発生しました。", http.StatusInternalServerError)
}

// back redirects to a same-app path given in the form (after a POST).
func (a *App) back(w http.ResponseWriter, r *http.Request, fallback string) {
	to := r.PostFormValue("back")
	if !strings.HasPrefix(to, a.base+"/") || strings.Contains(to, "//") || strings.ContainsAny(to, "\\\r\n") {
		to = fallback
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func formInt(r *http.Request, k string, min, max int) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue(k)))
	if err != nil || n < min || n > max {
		return 0, false
	}
	return n, true
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func limitForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return false
	}
	return true
}
