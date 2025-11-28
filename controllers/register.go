package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	beegoctx "github.com/beego/beego/v2/server/web/context"
	"golang.org/x/crypto/bcrypt"
)

type registerInput struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/auth/register
func AuthRegister(ctx *beegoctx.Context) {
	var in registerInput

	// Decode JSON body (with Beego fallback)
	body, _ := io.ReadAll(ctx.Request.Body)
	if len(body) == 0 && len(ctx.Input.RequestBody) > 0 {
		body = ctx.Input.RequestBody
	}
	if err := json.Unmarshal(body, &in); err != nil {
		jsonErr(ctx, http.StatusBadRequest, "invalid json")
		return
	}

	// Trim + validate
	in.FullName = strings.TrimSpace(in.FullName)
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)

	if msg := validateRegister(in); msg != "" {
		jsonErr(ctx, http.StatusBadRequest, msg)
		return
	}

	// Uniqueness checks
	exists, err := userExistsByEmail(in.Email)
	if err != nil {
		jsonErr(ctx, http.StatusInternalServerError, "server error")
		return
	}
	if exists {
		jsonErr(ctx, http.StatusConflict, "email already in use")
		return
	}
	exists, err = userExistsByUsername(in.Username)
	if err != nil {
		jsonErr(ctx, http.StatusInternalServerError, "server error")
		return
	}
	if exists {
		jsonErr(ctx, http.StatusConflict, "username already in use")
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonErr(ctx, http.StatusInternalServerError, "server error")
		return
	}

	// Insert user (use the same table name other controllers use)
	const q = "INSERT INTO " + usersTable + " (username, password_hash, full_name, email, role, create_at) VALUES (?,?,?,?,?, NOW())"
	res, err := srv.DB.Exec(q, in.Username, string(hash), in.FullName, in.Email, "user")
	if err != nil {
		// In case of a race, surface duplicate-key as 409
		lo := strings.ToLower(err.Error())
		if strings.Contains(lo, "duplicate") || strings.Contains(lo, "1062") {
			jsonErr(ctx, http.StatusConflict, "email or username already in use")
			return
		}
		log.Printf("register insert error: %v", err)
		jsonErr(ctx, http.StatusInternalServerError, "server error")
		return
	}
	newID, _ := res.LastInsertId()
	if newID <= 0 {
		jsonOK(ctx, map[string]interface{}{"status": "ok"})
		return
	}

	// Auto-login: issue a session token using the SAME machinery as login.go
	token := newToken()
	exp := time.Now().Add(sessionTTL()).UTC()
	sessionPut(Session{
		Token:  token,
		UserID: newID, // LastInsertId() is int64
		Email:  in.Email,
		Role:   "user",
		Exp:    exp, // matches Session struct in login.go
	})

	// Set auth cookie (SameSite=None; Secure for cross-origin over HTTPS)
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:     "imx_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Expires:  exp,
	})

	// Respond JSON
	jsonOK(ctx, map[string]interface{}{
		"token": token,
		"exp":   exp.Unix(),
		"user": map[string]interface{}{
			"id":        newID,
			"username":  in.Username,
			"full_name": in.FullName,
			"email":     in.Email,
			"role":      "user",
		},
	})
}

func validateRegister(in registerInput) string {
	if in.FullName == "" || in.Username == "" || in.Email == "" || in.Password == "" {
		return "please fill all fields"
	}
	if !strings.Contains(in.Email, "@") {
		return "invalid email"
	}
	if len(in.Username) < 3 {
		return "username must be at least 3 characters"
	}
	if len(in.Password) < 8 {
		return "password must be at least 8 characters"
	}
	return ""
}

func userExistsByEmail(email string) (bool, error) {
	var n int
	err := srv.DB.Get(&n, "SELECT COUNT(1) FROM "+usersTable+" WHERE email=?", email)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func userExistsByUsername(username string) (bool, error) {
	var n int
	err := srv.DB.Get(&n, "SELECT COUNT(1) FROM "+usersTable+" WHERE username=?", username)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
