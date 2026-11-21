package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	store, err := openStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.db.Close() })
	a, err := newApp(store, "/ap", "", strings.Repeat("s", 32), true)
	if err != nil {
		t.Fatal(err)
	}
	a.google = &GoogleConfig{ClientID: "cid", ClientSecret: "csecret", RedirectURL: "https://lab.example/ap/auth/callback", Allowed: []string{"me@example.com"}}
	return a
}

// stubToken serves an id_token with the given claims (nonce filled from the
// auth request when "nonce" is "<echo>").
func stubToken(t *testing.T, claims map[string]any, nonce *string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("client_secret") != "csecret" || r.PostForm.Get("code") != "good-code" || r.PostForm.Get("code_verifier") == "" {
			w.WriteHeader(400)
			return
		}
		c := map[string]any{}
		for k, v := range claims {
			c[k] = v
		}
		if c["nonce"] == "<echo>" {
			c["nonce"] = *nonce
		}
		payload, _ := json.Marshal(c)
		tok := "e30." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": tok})
	}))
	t.Cleanup(srv.Close)
	old := googleTokenURL
	googleTokenURL = srv.URL
	t.Cleanup(func() { googleTokenURL = old })
}

func login(t *testing.T, a *App, claims map[string]any, tamperState bool) (*httptest.ResponseRecorder, []*http.Cookie) {
	h := a.routes()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/ap/auth/google", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("start: %d", rec.Code)
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	q := loc.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("client_id") != "cid" {
		t.Fatalf("bad auth url: %s", loc)
	}
	nonce := q.Get("nonce")
	stubToken(t, claims, &nonce)
	state := q.Get("state")
	if tamperState {
		state = "x" + state
	}
	req := httptest.NewRequest("GET", "/ap/auth/callback?code=good-code&state="+url.QueryEscape(state), nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	return rec2, rec2.Result().Cookies()
}

func base() map[string]any {
	return map[string]any{"iss": "https://accounts.google.com", "aud": "cid", "exp": 4102444800, "nonce": "<echo>", "email": "Me@Example.com", "email_verified": true}
}

func sessionCookie(cs []*http.Cookie) *http.Cookie {
	for _, c := range cs {
		if c.Name == cookieName && c.Value != "" {
			return c
		}
	}
	return nil
}

func TestGoogleLoginSuccess(t *testing.T) {
	a := newTestApp(t)
	rec, cs := login(t, a, base(), false)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/ap/" {
		t.Fatalf("got %d %s", rec.Code, rec.Header().Get("Location"))
	}
	sc := sessionCookie(cs)
	if sc == nil || !sc.HttpOnly || !sc.Secure {
		t.Fatal("no secure session cookie")
	}
	req := httptest.NewRequest("GET", "/ap/plan", nil)
	req.AddCookie(sc)
	r := httptest.NewRecorder()
	a.routes().ServeHTTP(r, req)
	if r.Code != 200 {
		t.Fatalf("authed page: %d", r.Code)
	}
	// Removing the address from the allowlist revokes the session.
	a.google.Allowed = []string{"other@example.com"}
	r = httptest.NewRecorder()
	a.routes().ServeHTTP(r, req)
	if r.Code != http.StatusSeeOther {
		t.Fatalf("revoked session still works: %d", r.Code)
	}
}

func TestGoogleLoginRejects(t *testing.T) {
	cases := map[string]struct {
		mut    func(map[string]any)
		tamper bool
		want   string
	}{
		"not allowed": {func(c map[string]any) { c["email"] = "x@example.com" }, false, "e=denied"},
		"unverified":  {func(c map[string]any) { c["email_verified"] = false }, false, "e=denied"},
		"wrong aud":   {func(c map[string]any) { c["aud"] = "other" }, false, "e=failed"},
		"wrong iss":   {func(c map[string]any) { c["iss"] = "https://evil.example" }, false, "e=failed"},
		"expired":     {func(c map[string]any) { c["exp"] = 1 }, false, "e=failed"},
		"nonce":       {func(c map[string]any) { c["nonce"] = "nope" }, false, "e=failed"},
		"state":       {func(map[string]any) {}, true, "e=expired"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a := newTestApp(t)
			c := base()
			tc.mut(c)
			rec, cs := login(t, a, c, tc.tamper)
			if !strings.Contains(rec.Header().Get("Location"), tc.want) || sessionCookie(cs) != nil {
				t.Fatalf("got %d %s", rec.Code, rec.Header().Get("Location"))
			}
		})
	}
}

func TestPassphraseDisabledWithGoogle(t *testing.T) {
	a := newTestApp(t)
	req := httptest.NewRequest("POST", "/ap/login", strings.NewReader("password=anything"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("passphrase login should be disabled: %d", rec.Code)
	}
}
