package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv          string
	HTTPAddr        string
	MySQLDSN        string
	SessionTTLHours int // e.g. 168 for 7 days
}

// Load reads a simple KEY=VALUE config (comments starting with # ; //) and lets env vars override.
func Load(path string) (Config, error) {
	// defaults
	cfg := Config{
		AppEnv:          "dev",
		HTTPAddr:        ":8020",
		SessionTTLHours: 168, // 7 days
	}
	if path == "" {
		path = "conf/app.conf"
	}
	f, err := os.Open(filepath.Clean(path))
	if err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
				continue
			}
			if i := strings.IndexByte(line, '='); i > 0 {
				k := strings.TrimSpace(line[:i])
				v := strings.TrimSpace(line[i+1:])
				v = strings.Trim(v, `"'`)
				switch strings.ToUpper(k) {
				case "APP_ENV":
					cfg.AppEnv = v
				case "HTTP_ADDR":
					cfg.HTTPAddr = v
				case "MYSQL_DSN":
					cfg.MySQLDSN = v
				case "SESSION_TTL_HOURS":
					if n, _ := strconv.Atoi(v); n > 0 {
						cfg.SessionTTLHours = n
					}
				}
			}
		}
	}
	// env overrides (optional)
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.AppEnv = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		cfg.MySQLDSN = v
	}
	if v := os.Getenv("SESSION_TTL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.SessionTTLHours = n
		}
	}
	return cfg, nil
}
