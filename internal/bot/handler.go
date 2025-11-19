package bot

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

func (s *Service) AddRule(rule *BotRule) error {
	if rule.RuleType == "" || rule.Value == "" || rule.Action == "" {
		return fmt.Errorf("rule_type, value, and action are required")
	}
	return s.repo.Create(context.Background(), rule)
}

func (s *Service) ListRules() ([]*BotRule, error) {
	return s.repo.GetAll(context.Background())
}

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/bot/rules", h.Create)
	r.Get("/bot/rules", h.List)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var rule BotRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.AddRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.ListRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rules)
}
