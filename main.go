package main

import (
	"github.com/beego/beego/v2/server/web"
	"log"
	"vlu_infrastructure_management/controllers"
	_ "vlu_infrastructure_management/routers"
)

func main() {
	// Load config from cnf/app.conf
	if err := web.LoadAppConfig("ini", "conf/app.conf"); err != nil {
		log.Fatal(err)
	}

	//dsn, _ := web.AppConfig.String("MYSQL_DSN")
	//if dsn == "" {
	//	if d2, _ := web.AppConfig.String("mysql_dsn"); d2 != "" {
	//		dsn = d2
	//	}
	//}
	//if dsn == "" {
	//	log.Fatal("MYSQL_DSN not set in cnf/app.conf")
	//}
	//db, err := sqlx.Open("mysql", dsn)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//
	//const usersTable = "`vlu`.`users`" // set to "`user`" if DSN already selects vlu
	//
	//username := "admin"
	//fullName := "Administrator"
	//email := "admin@local"
	//role := "admin"
	//password := "Admin@12345"
	//
	//hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//
	//// requires UNIQUE on username or email
	//q := "INSERT INTO " + usersTable + " (username, full_name, email, password_hash, role, create_at) " +
	//	"VALUES (?, ?, ?, ?, ?, NOW()) " +
	//	"ON DUPLICATE KEY UPDATE full_name=VALUES(full_name), email=VALUES(email), password_hash=VALUES(password_hash), role=VALUES(role)"
	//if _, err := db.Exec(q, username, fullName, email, hash, role); err != nil {
	//	log.Fatalf("seed admin: %v", err)
	//}
	//
	//fmt.Println("✔ Admin ensured")
	//fmt.Println("  username:", username)
	//fmt.Println("  email   :", email)
	//fmt.Println("  password:", password)
	// Bootstrap Beego routes/filters/db/cache
	if _, err := controllers.Bootstrap(); err != nil {
		log.Fatal(err)
	}
	web.SetStaticPath("/static", "static")
	web.SetStaticPath("/img", "static/img")
	// Start the server
	web.Run() // uses HTTPAddr from app.conf if set
}
