package controllers

import (
	"log"
	"net/http"
	"strings"
	//"time"/

	beegoctx "github.com/beego/beego/v2/server/web/context"
)

const usersTable = "users"

// ---- token extraction (single return to keep other files unchanged) ----
func getToken(ctx *beegoctx.Context) string {
	// Authorization: Bearer <token>
	if auth := strings.TrimSpace(ctx.Input.Header("Authorization")); auth != "" {
		low := strings.ToLower(auth)
		if strings.HasPrefix(low, "bearer ") && len(auth) > 7 {
			return strings.TrimSpace(auth[7:])
		}
	}
	// Cookie: imx_token=<token>
	if c, err := ctx.Request.Cookie("imx_token"); err == nil && c != nil && c.Value != "" {
		return c.Value
	}
	// X-Auth-Token: <token>
	if h := strings.TrimSpace(ctx.Input.Header("X-Auth-Token")); h != "" {
		return h
	}
	// Query: ?token=<token> (useful for debugging/CLI)
	if q := strings.TrimSpace(ctx.Input.Query("token")); q != "" {
		return q
	}
	return ""
}

// small helper to expose user id to handlers that read different keys
func attachUser(ctx *beegoctx.Context, uid int64) {
	ctx.Input.SetData("user_id", uid)
	ctx.Input.SetData("userID", uid)
}

// Public (no auth): HTML/static, /api/healthz, /api/auth/*,
//
//	GET /api/items(/:id), /api/equipment-notes,
//	/api/instructions, /api/dashboard-stat
//
// Protected: everything else (POST/PUT/DELETE e.g. borrow/return/add item)
func SessionAuthFilter(ctx *beegoctx.Context) {
	path := ctx.Input.URL()
	method := strings.ToUpper(ctx.Input.Method())

	// Allow CORS preflights
	if method == http.MethodOptions {
		ctx.Output.Header("Access-Control-Allow-Origin", ctx.Input.Header("Origin"))
		ctx.Output.Header("Vary", "Origin")
		ctx.Output.Header("Access-Control-Allow-Credentials", "true")
		ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Auth-Token")
		ctx.Output.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
		return
	}

	// Non-API routes: serve SPA/static
	if !strings.HasPrefix(path, "/api/") {
		return
	}
	// Auth endpoints & health: always public
	if strings.HasPrefix(path, "/api/auth/") || path == "/api/healthz" {
		return
	}
	// Read-only public GETs (dashboard works for guests)
	if method == http.MethodGet {
		switch {
		case path == "/api/items",
			strings.HasPrefix(path, "/api/items/"),
			path == "/api/equipment-notes",
			path == "/api/instructions",
			path == "/api/dashboard-stat":
			return
		}
	}

	// Mutations: must be authenticated
	uid, ok := validateAndAttachUser(ctx)
	if !ok || uid <= 0 {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		_ = ctx.Output.JSON(map[string]any{"ok": false, "error": "unauthorized"}, false, false)
		return
	}
}

// Resolve user from token. Order:
//  1. in-memory session (sessionGet)
//  2. users table fallback: token/api_token/auth_token
func validateAndAttachUser(ctx *beegoctx.Context) (int64, bool) {
	tok := getToken(ctx)
	if tok == "" {
		log.Printf("[AUTH] 401 %s %s -> missing token", ctx.Input.Method(), ctx.Input.URL())
		return 0, false
	}

	// 1) in-memory session
	sess, ok := sessionGet(tok) // defined in login.go; returns your Session or nil
	if ok && sess != nil {
		switch s := sess.(type) {
		case Session:
			attachUser(ctx, s.UserID)
			log.Printf("[AUTH] ok via session user_id=%d path=%s", s.UserID, ctx.Input.URL())
			return s.UserID, true
		case map[string]interface{}:
			if v, ok := s["user_id"].(int64); ok && v > 0 {
				attachUser(ctx, v)
				log.Printf("[AUTH] ok via session(map) user_id=%d path=%s", v, ctx.Input.URL())
				return v, true
			}
		}
	}

	// 2) DB fallback (works after server restarts)
	srv := GetServer()
	if srv == nil || srv.DB == nil {
		log.Printf("[AUTH] 401 %s %s -> server/db nil", ctx.Input.Method(), ctx.Input.URL())
		return 0, false
	}

	// If you only have one column (e.g. token), keep the first predicate and drop the ORs.
	var uid int64
	err := srv.DB.Get(&uid, `
		SELECT id FROM `+usersTable+`
		WHERE token = ? OR api_token = ? OR auth_token = ?
		LIMIT 1
	`, tok, tok, tok)
	if err == nil && uid > 0 {
		attachUser(ctx, uid)
		log.Printf("[AUTH] ok via users table user_id=%d path=%s", uid, ctx.Input.URL())
		return uid, true
	}

	log.Printf("[AUTH] 401 %s %s -> token not found in session/users", ctx.Input.Method(), ctx.Input.URL())
	return 0, false
}
