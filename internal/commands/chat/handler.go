package chat

import (
	"log"

	"micheam.com/aico/internal/config"
)

// Handler handle the request and execute the command
type Handler struct {
	cfg    *config.Config
	logger *log.Logger
}

// New create a new Handler
func New(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}
