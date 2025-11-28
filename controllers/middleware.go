// middleware.go
package controllers

import (
	beegoctx "github.com/beego/beego/v2/server/web/context"
	"log"
	"net/http"
	"strings"
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

// VERY IMPORTANT: your login/register must bypass this filter.
// Also short-circuit CORS preflights.
func SessionAuthFilter(ctx *beegoctx.Context) {
	path := ctx.Input.URL()
	method := strings.ToUpper(ctx.Input.Method())

	// Always let OPTIONS through (CORS preflight)
	if method == http.MethodOptions {
		// Minimal CORS echo — real CORS is also configured in server.go via cors.Allow
		ctx.Output.Header("Access-Control-Allow-Origin", ctx.Input.Header("Origin"))
		ctx.Output.Header("Vary", "Origin")
		ctx.Output.Header("Access-Control-Allow-Credentials", "true")
		ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Auth-Token")
		ctx.Output.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
		return
	}

	// Public endpoints: health checks & auth
	if strings.HasPrefix(path, "/api/auth/") || path == "/api/healthz" {
		return
	}

	token := getToken(ctx)
	if token == "" {
		jsonErr(ctx, http.StatusUnauthorized, "missing token")
		ctx.ResponseWriter.Flush()
		return
	}

	// Validate token from in-memory session store
	sess, ok := sessionGet(token) // implemented in login.go
	if !ok {
		jsonErr(ctx, http.StatusUnauthorized, "invalid or expired token")
		ctx.ResponseWriter.Flush()
		return
	}

	// Stash a few fields for handlers to use
	switch s := sess.(type) {
	case Session:
		ctx.Input.SetData("userEmail", s.Email)
		ctx.Input.SetData("userID", s.UserID)
		ctx.Input.SetData("userRole", s.Role)
	case map[string]interface{}:
		if email, ok := s["email"].(string); ok {
			ctx.Input.SetData("userEmail", email)
		}
		if uid, ok := s["user_id"].(int64); ok {
			ctx.Input.SetData("userID", uid)
		}
		if role, ok := s["role"].(string); ok {
			ctx.Input.SetData("userRole", role)
		}
	}

	log.Printf("auth ok path=%s token=*** present", path)
}
