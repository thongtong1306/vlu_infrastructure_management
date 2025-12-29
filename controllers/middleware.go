package controllers

import (
	"net/http"
	"strings"

	beegoctx "github.com/beego/beego/v2/server/web/context"
)

const usersTable = "users"

func getToken(ctx *beegoctx.Context) string {
	// Authorization: Bearer <token>
	if auth := strings.TrimSpace(ctx.Input.Header("Authorization")); auth != "" {
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") && len(auth) > 7 {
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
	return ""
}

// SessionAuthFilter enforces auth for mutating APIs, but allows public read-only GETs.
// - Public (no auth): HTML/static, /api/healthz, /api/auth/*, GET of items/notes/instructions/dashboard-stat
// - Protected (auth required): POST/PUT/DELETE anywhere, and specific POST endpoints like borrow/return.
func SessionAuthFilter(ctx *beegoctx.Context) {
	p := ctx.Input.URL()
	m := ctx.Input.Method()

	// 1) Non-API routes (HTML pages, static files) are always public
	if !strings.HasPrefix(p, "/api/") {
		return
	}

	// 2) Auth endpoints & health are always public
	if strings.HasPrefix(p, "/api/auth/") || p == "/api/healthz" {
		return
	}

	// 3) Public read-only GET endpoints
	if m == http.MethodGet {
		switch {
		case p == "/api/items",
			strings.HasPrefix(p, "/api/items/"), // e.g. /api/items/:id
			p == "/api/equipment-notes",
			p == "/api/instructions",
			p == "/api/dashboard-stat":
			return
		}
	}

	// 4) Everything else requires a valid session/token
	uid, ok := validateAndAttachUser(ctx) // sets both "user_id" and "userID" on success
	if !ok || uid <= 0 {
		ctx.Output.SetStatus(401)
		_ = ctx.Output.JSON(map[string]any{"ok": false, "error": "unauthorized"}, false, false)
		return
	}
}

// validateAndAttachUser validates your cookie/Bearer token and sets user on context.
// Implement this according to your existing session logic.
func validateAndAttachUser(ctx *beegoctx.Context) (int64, bool) {
	// ... your existing auth code (cookie or Authorization: Bearer ...)
	// suppose you get uid int64 on success:
	// ctx.Input.SetData("user_id", uid)
	// ctx.Input.SetData("userID", uid)
	// return uid, true

	return 0, false // placeholder if you already have this implemented elsewhere
}
