// controllers/api.go
package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	"sort"
	"strconv"
	"strings"
	"vlu_infrastructure_management/models"
)

type Api struct {
	web.Controller
}

type EquipListResp struct {
	Rows   []models.Equipment `json:"rows"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
	Total  int                `json:"total"`
	Order  string             `json:"order"`
}

func normalizeRowMap(m map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		if b, ok := v.([]byte); ok {
			m[k] = string(b)
		}
	}
	return m
}

var tableNameToSQL = map[string]string{
	"log_lab_equipment_master":    "`log_lab_equipment_master`",
	"log_lab_borrow_records":      "`log_lab_borrow_records`",
	"log_lab_calibration_logs":    "`log_lab_calibration_logs`",
	"log_lab_maintenance_records": "`log_lab_maintenance_records`",
	"log_lab_activity_logs":       "`log_lab_activity_logs`",
	"log_lab_storage":             "`log_lab_storage`",
}

var tableOrder = map[string]string{
	"log_lab_equipment_master":    "id DESC",
	"log_lab_borrow_records":      "borrow_date DESC, id DESC",
	"log_lab_calibration_logs":    "`date` DESC, id DESC",
	"log_lab_maintenance_records": "date_reported DESC, id DESC",
	"log_lab_activity_logs":       "`timestamp` DESC, id DESC",
	"log_lab_storage":             "id DESC",
}

// GET /api/equipment?limit=200&offset=0&order_by=id&order_dir=desc
func (c *Api) GetAllEquipment() {
	raw := strings.TrimSpace(c.GetString("tables"))
	var names []string
	if raw == "" || raw == "*" {
		for k := range tableNameToSQL {
			names = append(names, k)
		}
		sort.Strings(names) // stable output order
	} else {
		for _, p := range strings.Split(raw, ",") {
			if n := strings.TrimSpace(p); n != "" {
				names = append(names, n)
			}
		}
	}

	// Pagination for each table
	limit := 200
	if v := c.GetString("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	offset := 0
	if v := c.GetString("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	result := make(map[string]interface{}, len(names)+1)
	invalid := []string{}

	for _, name := range names {
		sqlName, ok := tableNameToSQL[name]
		if !ok {
			invalid = append(invalid, name)
			continue
		}
		order := tableOrder[name]
		if order == "" {
			order = "id DESC"
		}

		q := fmt.Sprintf("SELECT * FROM %s ORDER BY %s LIMIT ? OFFSET ?", sqlName, order)
		rows, err := srv.DB.Queryx(q, limit, offset)
		if err != nil {
			result[name] = map[string]interface{}{"error": "query error"}
			continue
		}
		func() {
			defer rows.Close()
			list := make([]map[string]interface{}, 0, limit)
			for rows.Next() {
				m := map[string]interface{}{}
				if err := rows.MapScan(m); err != nil {
					continue
				}
				list = append(list, normalizeRowMap(m))
			}
			result[name] = list
		}()
	}

	if len(invalid) > 0 {
		result["_invalid"] = invalid
	}
	// Optional meta if you want:
	// result["_meta"] = map[string]interface{}{"limit": limit, "offset": offset}

	c.Data["json"] = result
	c.ServeJSON()
}
