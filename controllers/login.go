package controllers

import (
	cryptoRand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"vlu_infrastructure_management/models"

	beego "github.com/beego/beego/v2/server/web"
	beegoctx "github.com/beego/beego/v2/server/web/context"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Token  string
	UserID int64
	Email  string
	Role   string
	Exp    time.Time
}

// ---- simple in-memory session store via patrickmn/go-cache (wired in server.go) ----
func sessionTTL() time.Duration {
	h, _ := beego.AppConfig.Int64("SESSION_TTL_HOURS")
	if h <= 0 {
		h = 168
	}
	return time.Duration(h) * time.Hour
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = cryptoRand.Read(b)
	return hex.EncodeToString(b)
}

// Backed by server.go’s cache instance:
func sessionPut(s Session) {
	srv := GetServer() // see server.go; returns the singleton
	srv.Cache.Set(s.Token, s, sessionTTL())
}
func sessionGet(token string) (interface{}, bool) {
	srv := GetServer()
	return srv.Cache.Get(token)
}
func sessionDel(token string) {
	srv := GetServer()
	srv.Cache.Delete(token)
}

type loginInput struct {
	Identifier string `json:"identifier"` // username or email
	Password   string `json:"password"`
}

func AuthLogin(ctx *beegoctx.Context) {
	// Read body
	body, _ := io.ReadAll(ctx.Request.Body)
	if len(body) == 0 && len(ctx.Input.RequestBody) > 0 {
		body = ctx.Input.RequestBody
	}

	var in loginInput
	if err := json.Unmarshal(body, &in); err != nil {
		jsonErr(ctx, http.StatusBadRequest, "invalid json")
		return
	}
	id := strings.TrimSpace(in.Identifier)
	pw := in.Password
	if id == "" || pw == "" {
		jsonErr(ctx, http.StatusBadRequest, "missing identifier or password")
		return
	}

	// Look up user (either by email or username)
	srv := GetServer()
	var u models.User
	q := `SELECT id, username, password_hash, email, role FROM users WHERE email = ? OR username = ? LIMIT 1`
	if err := srv.DB.Get(&u, q, id, id); err != nil {
		if err == sql.ErrNoRows {
			jsonErr(ctx, http.StatusUnauthorized, "invalid credentials")
		} else {
			log.Println("login query error:", err)
			jsonErr(ctx, http.StatusInternalServerError, "server error")
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(pw)); err != nil {
		jsonErr(ctx, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Build a session & token
	token := newToken()
	exp := time.Now().Add(sessionTTL())
	sessionPut(Session{
		Token:  token,
		UserID: int64(u.ID),
		Email:  u.Email,
		Role:   u.Role,
		Exp:    exp,
	})

	// Optionally also set a cookie so non-Redux clients can work
	// If you host frontend and backend on different origins, use SameSite=None; Secure
	cookie := &http.Cookie{
		Name:     "imx_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		// Comment the next 2 lines if same-origin; keep if cross-origin over HTTPS
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Expires:  exp,
	}
	http.SetCookie(ctx.ResponseWriter, cookie)

	// Reply JSON for the React client (which reads `data.token`)
	jsonOK(ctx, map[string]interface{}{
		"token": token,
		"exp":   exp.Unix(),
		"user": map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role,
		},
	})
}

func AuthLogout(ctx *beegoctx.Context) {
	// Invalidate session if present
	if t := getToken(ctx); t != "" {
		sessionDel(t)
	}
	// Clear cookie
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:     "imx_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})
	jsonOK(ctx, map[string]interface{}{"ok": true})
}
