// controllers/server.go
package controllers

import (
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"

	beego "github.com/beego/beego/v2/server/web"
	beegoctx "github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

type Server struct {
	DB    *sqlx.DB
	Cache *cache.Cache
}

var srv *Server

// GetServer exposes the singleton to other controllers (used by login/session helpers).
func GetServer() *Server {
	return srv
}

func Bootstrap() (*Server, error) {
	return InitServer()
}

// InitServer initializes DB, cache, CORS, and a tiny health endpoint.
// Call this once in main.go before beego.Run():
//
//	if _, err := controllers.InitServer(); err != nil { log.Fatal(err) }
func InitServer() (*Server, error) {
	// ---- 1) Load configuration
	mysqlDSN := firstNonEmpty(
		getConf("mysql_dsn"),
		getConf("db_dsn"),
		os.Getenv("MYSQL_DSN"),
		os.Getenv("DB_DSN"),
	)
	if mysqlDSN == "" {
		// Common local default:
		//   user:pass@tcp(127.0.0.1:3306)/vlu?parseTime=true&charset=utf8mb4,utf8
		mysqlDSN = "root:root@tcp(127.0.0.1:3306)/vlu_infrastructure?parseTime=true&charset=utf8mb4,utf8"
		log.Printf("[server] MYSQL_DSN not set; using default DSN: %s", mysqlDSN)
	}

	// Session TTL (hours)
	sessTTL := int64FromConf("SESSION_TTL_HOURS", 168) // 7 days

	// Allowed CORS origins (comma-separated)
	originsCSV := firstNonEmpty(
		getConf("cors_origins"),
		os.Getenv("CORS_ORIGINS"),
		"http://localhost:5173,http://localhost:8020",
	)
	allowOrigins := splitCSV(originsCSV)

	// ---- 2) Open DB (sqlx + MySQL)
	db, err := sqlx.Open("mysql", mysqlDSN)
	if err != nil {
		return nil, err
	}
	// Reasonable pool settings for a small app; tune as needed
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(15 * time.Minute)
	db.SetConnMaxLifetime(2 * time.Hour)

	// Ping to verify connectivity now (fail fast)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// ---- 3) In-memory cache for sessions (used by login.go helpers)
	defaultExp := time.Duration(sessTTL) * time.Hour
	srvCache := cache.New(defaultExp, 10*time.Minute)

	// ---- 4) CORS (allow credentials, common headers & methods)
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		// Keep this false for security; enumerate explicit origins instead.
		AllowAllOrigins: false,
		AllowOrigins:    allowOrigins,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "X-Auth-Token"},
		ExposeHeaders:   []string{"Content-Length", "Content-Type"},
		// Enable if your frontend uses cookies for auth; safe with explicit origins.
		AllowCredentials: true,
		// Optional: MaxAge caches preflight for 1 hour
		MaxAge: 3600,
	}))

	// ---- 5) Health endpoint (unauthenticated)
	beego.Get("/api/healthz", func(ctx *beegoctx.Context) {
		ctx.Output.Header("Cache-Control", "no-store")
		ctx.Output.JSON(map[string]interface{}{
			"ok":        true,
			"service":   "vlu_infrastructure_management",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}, false, false)
	})

	// ---- 6) Save singleton
	srv = &Server{
		DB:    db,
		Cache: srvCache,
	}

	log.Printf("[server] Init OK | origins=%v | sessionTTL=%dh", allowOrigins, sessTTL)
	return srv, nil
}

// ---------- helpers ----------

func getConf(key string) string {
	// Reads from app.conf (if present). Returns empty string on error/missing.
	v, err := beego.AppConfig.String(key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}

func int64FromConf(key string, def int64) int64 {
	// env first
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		if n, err := parseInt64(s); err == nil {
			return n
		}
	}
	// app.conf
	if s := getConf(key); s != "" {
		if n, err := parseInt64(s); err == nil {
			return n
		}
	}
	return def
}

func parseInt64(s string) (int64, error) {
	var n int64
	var sign int64 = 1
	if s != "" && (s[0] == '-' || s[0] == '+') {
		if s[0] == '-' {
			sign = -1
		}
		s = s[1:]
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, beego.ErrAbort
		}
		n = n*10 + int64(s[i]-'0')
	}
	return sign * n, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
