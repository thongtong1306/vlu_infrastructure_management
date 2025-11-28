package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

const equipNotesTable = "log_lab_equipment_notes"

type EquipNote struct {
	ID        int64      `db:"id"         json:"id"`
	ItemID    int64      `db:"item_id"    json:"item_id"`
	NoteText  string     `db:"note_text"  json:"note_text"`
	CreatedBy *string    `db:"created_by" json:"created_by,omitempty"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
}

type EquipmentNoteController struct{ web.Controller }

// GET /api/equipment-notes?item_id=123
func (c *EquipmentNoteController) GetByItem() {
	itemIDStr := strings.TrimSpace(c.GetString("item_id"))
	if itemIDStr == "" {
		jsonErr(c.Ctx, http.StatusBadRequest, "item_id is required")
		return
	}
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil || itemID <= 0 {
		jsonErr(c.Ctx, http.StatusBadRequest, "invalid item_id")
		return
	}
	s := GetServer()
	var rows []EquipNote
	q := "SELECT id, item_id, note_text, created_by, created_at FROM " + equipNotesTable + " WHERE item_id=? ORDER BY id DESC"
	if err := s.DB.Select(&rows, q, itemID); err != nil {
		jsonErr(c.Ctx, http.StatusInternalServerError, "query error")
		return
	}
	jsonOK(c.Ctx, rows)
}

// POST /api/equipment-notes
func (c *EquipmentNoteController) Add() {
	var in struct {
		ItemID    int64   `json:"item_id"`
		NoteText  string  `json:"note_text"`
		CreatedBy *string `json:"created_by"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		jsonErr(c.Ctx, http.StatusBadRequest, "invalid json")
		return
	}
	if in.ItemID <= 0 || strings.TrimSpace(in.NoteText) == "" {
		jsonErr(c.Ctx, http.StatusBadRequest, "item_id and note_text are required")
		return
	}
	s := GetServer()
	if _, err := s.DB.Exec(
		"INSERT INTO "+equipNotesTable+" (item_id, note_text, created_by) VALUES (?,?,?)",
		in.ItemID, strings.TrimSpace(in.NoteText), in.CreatedBy,
	); err != nil {
		jsonErr(c.Ctx, http.StatusInternalServerError, "insert error")
		return
	}
	jsonOK(c.Ctx, map[string]interface{}{"ok": true})
}
