package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Google sign-in: OpenID Connect authorization code flow with PKCE, state and
// nonce. The ID token comes straight from Google's token endpoint over TLS in
// exchange for our client secret, so per OIDC Core 3.1.3.7 its signature does
// not need separate verification; we still check iss, aud, exp, nonce,
// email_verified and the email allowlist.

const (
	googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"

	oauthCookie = "ap_oauth"
	oauthTTL    = 30 * time.Minute // time allowed between "Googleでログイン" and the callback
)

type GoogleConfig struct {
	ClientID, ClientSecret, RedirectURL string
	Allowed                             []string // lower-cased emails
}

func (g *GoogleConfig) enabled() bool { return g != nil && g.ClientID != "" }

func randToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (a *App) mac(parts ...string) string {
	m := hmac.New(sha256.New, a.secret)
	m.Write([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(m.Sum(nil))
}

// ---- session cookie: v2.<email b64>.<exp>.<sig> ----

func (a *App) setSession(w http.ResponseWriter, email string) {
	e64 := base64.RawURLEncoding.EncodeToString([]byte(email))
	exp := strconv.FormatInt(time.Now().Add(sessionTTL).Unix(), 10)
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: "v2." + e64 + "." + exp + "." + a.mac("session", e64, exp), Path: a.base + "/",
		MaxAge: int(sessionTTL.Seconds()), HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode,
	})
}

// sessionEmail returns the signed-in identity, re-checking the allowlist so
// removing an address from ALLOWED_EMAILS revokes its sessions.
func (a *App) sessionEmail(r *http.Request) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 4 || parts[0] != "v2" || !hmac.Equal([]byte(parts[3]), []byte(a.mac("session", parts[1], parts[2]))) {
		return "", false
	}
	exp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Now().Unix() >= exp {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	email := string(raw)
	if a.google.enabled() {
		return email, slices.Contains(a.google.Allowed, email)
	}
	return email, email == passphraseIdentity
}

const passphraseIdentity = "passphrase"

// ---- handlers ----

func (a *App) googleStart(w http.ResponseWriter, r *http.Request) {
	if !a.google.enabled() {
		http.NotFound(w, r)
		return
	}
	state, nonce, verifier := randToken(24), randToken(24), randToken(48)
	exp := strconv.FormatInt(time.Now().Add(oauthTTL).Unix(), 10)
	val := strings.Join([]string{state, nonce, verifier, exp}, ".")
	http.SetCookie(w, &http.Cookie{
		Name: oauthCookie, Value: val + "." + a.mac("oauth", val), Path: a.base + "/auth/",
		MaxAge: int(oauthTTL.Seconds()), HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode,
	})
	sum := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"client_id":             {a.google.ClientID},
		"redirect_uri":          {a.google.RedirectURL},
		"response_type":         {"code"},
		"scope":                 {"openid email"},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
		"prompt":                {"select_account"},
	}
	http.Redirect(w, r, googleAuthURL+"?"+q.Encode(), http.StatusFound)
}

func (a *App) loginError(w http.ResponseWriter, r *http.Request, code, logMsg string, args ...any) {
	slog.Warn("google_login_failed", append([]any{"reason", logMsg}, args...)...)
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Value: "", Path: a.base + "/auth/", MaxAge: -1, HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, a.base+"/login?e="+code, http.StatusSeeOther)
}

func (a *App) googleCallback(w http.ResponseWriter, r *http.Request) {
	if !a.google.enabled() {
		http.NotFound(w, r)
		return
	}
	c, err := r.Cookie(oauthCookie)
	if err != nil {
		a.loginError(w, r, "expired", "no oauth cookie")
		return
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 5 || !hmac.Equal([]byte(parts[4]), []byte(a.mac("oauth", strings.Join(parts[:4], ".")))) {
		a.loginError(w, r, "expired", "bad oauth cookie")
		return
	}
	state, nonce, verifier := parts[0], parts[1], parts[2]
	if exp, err := strconv.ParseInt(parts[3], 10, 64); err != nil || time.Now().Unix() >= exp {
		a.loginError(w, r, "expired", "oauth cookie expired")
		return
	}
	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		a.loginError(w, r, "cancel", "google returned error", "error", e)
		return
	}
	if !hmac.Equal([]byte(q.Get("state")), []byte(state)) {
		a.loginError(w, r, "expired", "state mismatch")
		return
	}
	claims, err := a.exchangeCode(r, q.Get("code"), verifier)
	if err != nil {
		a.loginError(w, r, "failed", "token exchange", "err", err)
		return
	}
	email := strings.ToLower(claims.Email)
	switch {
	case claims.Iss != "https://accounts.google.com" && claims.Iss != "accounts.google.com":
		a.loginError(w, r, "failed", "bad iss")
	case !claims.audIs(a.google.ClientID):
		a.loginError(w, r, "failed", "bad aud")
	case time.Now().Unix() >= claims.Exp:
		a.loginError(w, r, "failed", "token expired")
	case !hmac.Equal([]byte(claims.Nonce), []byte(nonce)):
		a.loginError(w, r, "failed", "nonce mismatch")
	case !claims.verified():
		a.loginError(w, r, "denied", "email not verified")
	case !slices.Contains(a.google.Allowed, email):
		a.loginError(w, r, "denied", "email not allowed")
	default:
		http.SetCookie(w, &http.Cookie{Name: oauthCookie, Value: "", Path: a.base + "/auth/", MaxAge: -1, HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode})
		a.setSession(w, email)
		slog.Info("login", "method", "google")
		http.Redirect(w, r, a.base+"/", http.StatusSeeOther)
	}
}

type idClaims struct {
	Iss           string          `json:"iss"`
	Aud           json.RawMessage `json:"aud"`
	Exp           int64           `json:"exp"`
	Nonce         string          `json:"nonce"`
	Email         string          `json:"email"`
	EmailVerified any             `json:"email_verified"`
}

func (c idClaims) audIs(id string) bool {
	var one string
	if json.Unmarshal(c.Aud, &one) == nil {
		return one == id
	}
	var many []string
	return json.Unmarshal(c.Aud, &many) == nil && len(many) == 1 && many[0] == id
}

func (c idClaims) verified() bool {
	switch v := c.EmailVerified.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	}
	return false
}

// googleTokenURL is a var so tests can point it at a stub.
var googleTokenURL = "https://oauth2.googleapis.com/token"

var tokenClient = &http.Client{Timeout: 10 * time.Second}

func (a *App) exchangeCode(r *http.Request, code, verifier string) (idClaims, error) {
	var claims idClaims
	if code == "" {
		return claims, fmt.Errorf("missing code")
	}
	form := url.Values{
		"code":          {code},
		"client_id":     {a.google.ClientID},
		"client_secret": {a.google.ClientSecret},
		"redirect_uri":  {a.google.RedirectURL},
		"grant_type":    {"authorization_code"},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return claims, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := tokenClient.Do(req)
	if err != nil {
		return claims, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if err != nil {
		return claims, err
	}
	if res.StatusCode != http.StatusOK {
		return claims, fmt.Errorf("token endpoint status %d", res.StatusCode)
	}
	var tok struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return claims, err
	}
	seg := strings.Split(tok.IDToken, ".")
	if len(seg) != 3 {
		return claims, fmt.Errorf("malformed id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(seg[1])
	if err != nil {
		return claims, err
	}
	return claims, json.Unmarshal(payload, &claims)
}
