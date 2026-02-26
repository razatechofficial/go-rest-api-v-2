package container

import (
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/order"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/user"
)

// Handlers holds all HTTP delivery layer implementations.
type Handlers struct {
	User  *user.Handler
	Order *order.Handler
}

func (c *Container) buildHandlers() *Handlers {
	return &Handlers{
		User:  user.NewHandler(c.Services.User),
		Order: order.NewHandler(c.Services.Order),
	}
}
