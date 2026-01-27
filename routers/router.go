package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"vlu_infrastructure_management/controllers"
)

func init() {
	// ----- API (first) -----
	beego.InsertFilter("/api/*", beego.BeforeRouter, controllers.SessionAuthFilter)
	beego.Post("/api/auth/login", controllers.AuthLogin)
	beego.Post("/api/auth/logout", controllers.AuthLogout)
	beego.Post("/api/auth/register", controllers.AuthRegister)
	beego.Router("/api/dashboard-stat", &controllers.Api{}, "get:GetAllEquipment")
	beego.Router("/api/items", &controllers.ItemController{}, "get:GetAll;post:Add")
	beego.Router("/api/items/borrow", &controllers.ItemController{}, "post:Borrow")
	beego.Router("/api/items/return", &controllers.ItemController{}, "post:Return")
	beego.Router("/api/instructions", &controllers.InstructionController{}, "get:GetByItem;post:Add")
	beego.Router("/api/instructions/:id([0-9]+)", &controllers.InstructionController{}, "get:GetOne")
	beego.Router("/api/equipment-notes", &controllers.EquipmentNoteController{}, "get:GetByItem;post:Add")
	beego.Router("/api/items/:id([0-9]+)/image", &controllers.ItemController{}, "put:UpdateImageURL")
	beego.Router("/api/items/:id([0-9]+)", &controllers.ItemController{}, "get:GetOne")
	beego.Router("/api/items/open-borrows", &controllers.ItemController{}, "get:GetOpenBorrows")

	// ----- Pages (SPA shell) -----
	beego.Router("/", &controllers.MainController{}, "get:Home")     // server decides: /dashboard or /login
	beego.Router("/login", &controllers.MainController{}, "get:Get") // SPA handles /login view
	beego.Router("/dashboard", &controllers.MainController{}, "get:Get")
	beego.Router("/labs/*", &controllers.MainController{}, "get:Get") // e.g., /labs/lab-3
	beego.Router("/equipments", &controllers.MainController{}, "get:Get")
	beego.Router("/instructions/*", &controllers.MainController{}, "get:Get")
}
