// Package notes adalah aplikasi demo CRUD in-memory: bukti M2 bahwa
// provider aplikasi bisa mendaftarkan route ke core/httpx tanpa menyentuh
// kernel. Penyimpanan mati bersama proses.
package notes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/sonix-framework/core/httpx"
)

// Note adalah resource demo.
type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// noteIn adalah payload create/update dari klien.
type noteIn struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Service menyimpan notes di memori dengan akses terlindungi mutex.
type Service struct {
	mu     sync.RWMutex
	notes  map[int]Note
	nextID int
}

// NewService membuat penyimpanan notes kosong.
func NewService() *Service {
	return &Service{notes: map[int]Note{}, nextID: 1}
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
func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "id harus angka")
		return 0, false
	}
	return id, true
}

// list melayani GET /notes: seluruh notes terurut id naik.
func (a *API) list(w http.ResponseWriter, r *http.Request) {
	a.svc.mu.RLock()
	defer a.svc.mu.RUnlock()
	out := make([]Note, 0, len(a.svc.notes))
	for id := 1; id < a.svc.nextID; id++ {
		if n, ok := a.svc.notes[id]; ok {
			out = append(out, n)
		}
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
	a.svc.mu.Lock()
	defer a.svc.mu.Unlock()
	n := Note{ID: a.svc.nextID, Title: in.Title, Done: in.Done}
	a.svc.notes[n.ID] = n
	a.svc.nextID++
	writeJSON(w, http.StatusCreated, n)
}

// get melayani GET /notes/{id}: 200 atau 404.
func (a *API) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a.svc.mu.RLock()
	defer a.svc.mu.RUnlock()
	n, found := a.svc.notes[id]
	if !found {
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
	a.svc.mu.Lock()
	defer a.svc.mu.Unlock()
	if _, found := a.svc.notes[id]; !found {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "catatan tidak ditemukan")
		return
	}
	n := Note{ID: id, Title: in.Title, Done: in.Done}
	a.svc.notes[id] = n
	writeJSON(w, http.StatusOK, n)
}

// remove melayani DELETE /notes/{id}: 204 atau 404.
func (a *API) remove(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a.svc.mu.Lock()
	defer a.svc.mu.Unlock()
	if _, found := a.svc.notes[id]; !found {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "catatan tidak ditemukan")
		return
	}
	delete(a.svc.notes, id)
	w.WriteHeader(http.StatusNoContent)
}
