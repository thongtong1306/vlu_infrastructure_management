package controllers

import (
	"github.com/beego/beego/v2/server/web"
)

type MainController struct {
	web.Controller
}

// add to MainController
func (c *MainController) Home() {
	// Try header/cookie (server can’t see localStorage)
	token := getToken(c.Ctx)
	if token == "" {
		if ck, err := c.Ctx.Request.Cookie("imx_token"); err == nil && ck != nil && ck.Value != "" {
			token = ck.Value
		}
	}
	if token != "" {
		if _, ok := sessionGet(token); ok {
			c.Redirect("/dashboard", 302)
			return
		}
	}
	c.Redirect("/login", 302)
}

// GET /*  -> serve SPA shell so client-router can render /labs, /items, etc.
func (c *MainController) Get() {
	c.Data["Website"] = "vanlang-infrastructure.org"
	c.Data["Email"] = "thongtong.iuevent@gmail.com"
	// This should be your SPA entry template (unchanged)
	c.TplName = "infrastructure-manage.html"
}
