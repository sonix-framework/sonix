// Package notes adalah aplikasi demo CRUD berbasis ORM: bukti M4 bahwa
// provider aplikasi bisa memakai core/orm di atas core/database. Tabel
// dibuat otomatis saat start (ensureSchema), data bertahan antar restart.
package notes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sonix-framework/core/database"
	"github.com/sonix-framework/core/httpx"
	"github.com/sonix-framework/core/orm"
)

// Note adalah resource demo; db:"id" menjadikannya primary key yang
// diisi otomatis oleh Insert (RETURNING id / LastInsertId).
type Note struct {
	ID        int64     `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Done      bool      `db:"done" json:"done"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// noteIn adalah payload create/update dari klien.
type noteIn struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Service menyimpan notes lewat ORM di database terkonfigurasi.
type Service struct {
	db *database.DB
}

// NewService membuka akses ORM dan memastikan tabel notes ada.
func NewService(db *database.DB) (*Service, error) {
	if err := ensureSchema(db); err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

// ensureSchema membuat tabel notes bila belum ada, dengan DDL sesuai dialek.
func ensureSchema(db *database.DB) error {
	var ddl string
	switch db.Dialect().Name() {
	case "postgres":
		ddl = `CREATE TABLE IF NOT EXISTS notes (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ)`
	case "mysql":
		ddl = `CREATE TABLE IF NOT EXISTS notes (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			done BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL)`
	default: // sqlite
		ddl = `CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT 0,
			created_at TIMESTAMP,
			updated_at TIMESTAMP)`
	}
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("notes: siapkan tabel: %w", err)
	}
	return nil
}

// API menghubungkan Service ke HTTP.
type API struct {
	svc *Service
}

// NewAPI membangun API di atas Service.
func NewAPI(svc *Service) *API {
	return &API{svc: svc}
}

// writeJSON menuliskan response JSON dengan header konsisten.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// pathID membaca {id} dari path; menulis error langsung bila bukan angka.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "id harus angka")
		return 0, false
	}
	return id, true
}

// list melayani GET /notes: seluruh notes terurut id naik.
func (a *API) list(w http.ResponseWriter, r *http.Request) {
	out, err := orm.Query[Note](a.svc.db).OrderBy("id").Get(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// create melayani POST /notes: 201 dengan Note baru.
func (a *API) create(w http.ResponseWriter, r *http.Request) {
	var in noteIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "body JSON tidak valid")
		return
	}
	if in.Title == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid_title", "title tidak boleh kosong")
		return
	}
	n := Note{Title: in.Title, Done: in.Done}
	if err := orm.Query[Note](a.svc.db).Insert(r.Context(), &n); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

// get melayani GET /notes/{id}: 200 atau 404.
func (a *API) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	n, err := orm.Query[Note](a.svc.db).Find(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "catatan tidak ditemukan")
		return
	}
	writeJSON(w, http.StatusOK, n)
}

// update melayani PUT /notes/{id}: replace penuh, 200 atau 404.
func (a *API) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in noteIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "body JSON tidak valid")
		return
	}
	if in.Title == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid_title", "title tidak boleh kosong")
		return
	}
	n, err := orm.Query[Note](a.svc.db).Where("id", id).Update(r.Context(),
		map[string]any{"title": in.Title, "done": in.Done})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	if n == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "catatan tidak ditemukan")
		return
	}
	out, err := orm.Query[Note](a.svc.db).Find(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "reload_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// remove melayani DELETE /notes/{id}: 204 atau 404.
func (a *API) remove(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	n, err := orm.Query[Note](a.svc.db).Where("id", id).Delete(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	if n == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "catatan tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
