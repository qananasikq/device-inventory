package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"device-inventory/internal/repository"
)

const maxHostnameLen = 64

type apiError struct {
	Error string `json:"error"`
}

type DeviceHandler struct {
	repo *repository.DeviceRepo
}

func NewDeviceHandler(repo *repository.DeviceRepo) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d repository.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid json body"})
		return
	}
	if msg := validateDeviceInput(d); msg != "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: msg})
		return
	}

	if err := h.repo.Create(&d); err != nil {
		log.Printf("create device error: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	log.Printf("created device id=%d hostname=%s", d.ID, d.Hostname)
	writeJSON(w, http.StatusCreated, d)
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	var activeFilter *bool
	if v := r.URL.Query().Get("is_active"); v != "" {
		switch v {
		case "true":
			b := true
			activeFilter = &b
		case "false":
			b := false
			activeFilter = &b
		default:
			writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid is_active value"})
			return
		}
	}

	search := r.URL.Query().Get("search")

	devices, err := h.repo.GetAll(activeFilter, search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

func (h *DeviceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	device, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrDeviceNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "device not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var d repository.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid json body"})
		return
	}
	if msg := validateDeviceInput(d); msg != "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: msg})
		return
	}

	updated, err := h.repo.Update(id, d)
	if err != nil {
		if errors.Is(err, repository.ErrDeviceNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "device not found"})
			return
		}
		log.Printf("update device id=%s error: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	log.Printf("updated device id=%s", id)
	writeJSON(w, http.StatusOK, updated)
}

// soft delete
func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Deactivate(id); err != nil {
		if errors.Is(err, repository.ErrDeviceNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "device not found"})
			return
		}
		log.Printf("deactivate device id=%s error: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}
	log.Printf("deactivated device id=%s", id)
	w.WriteHeader(http.StatusNoContent)
}

// проверка перед записью
func validateDeviceInput(d repository.Device) string {
	hostname := strings.TrimSpace(d.Hostname)
	if hostname == "" {
		return "hostname is required"
	}
	if len(hostname) > maxHostnameLen {
		return "hostname is too long"
	}

	ip := net.ParseIP(strings.TrimSpace(d.IP))
	if ip == nil || ip.To4() == nil {
		return "invalid ip address"
	}

	if strings.TrimSpace(d.Location) == "" {
		return "location is required"
	}

	return ""
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
