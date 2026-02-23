// internal/container/container.go
package container

import (
	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"

	"github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"
)

// Container is the composition root of the application.
// It is the only struct that knows about all modules.
// Built once at startup, lives for the entire process lifetime.
type Container struct {
	cfg *config.Config
	DB  *postgres.Pool

	Repositories *Repositories
	Services     *Services
	Handlers     *Handlers
}

// New builds the entire dependency graph in the correct order.
// Repositories → Services → Handlers
// Each layer depends only on the layer below it.
func New(cfg *config.Config, db *postgres.Pool) *Container {
	c := &Container{
		cfg: cfg,
		DB:  db,
	}

	c.Repositories = c.buildRepositories()
	c.Services = c.buildServices()
	c.Handlers = c.buildHandlers()

	logger.Info("container initialized successfully")

	return c
}
