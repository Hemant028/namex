package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"


	"github.com/go-chi/chi/v5"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDomain(d *Domain) error {
	// Basic validation
	if d.Name == "" || d.TargetURL == "" {
		return fmt.Errorf("name and target_url are required")
	}
	if d.Config == nil {
		d.Config = json.RawMessage("{}")
	}
	return s.repo.Create(context.Background(), d)
}

func (s *Service) GetDomain(name string) (*Domain, error) {
	return s.repo.GetByName(context.Background(), name)
}

func (s *Service) ListDomains() ([]*Domain, error) {
	return s.repo.GetAll(context.Background())
}

// Handler
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/domains", h.Create)
	r.Get("/domains", h.List)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var d Domain
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.CreateDomain(&d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	domains, err := h.service.ListDomains()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(domains)
}
