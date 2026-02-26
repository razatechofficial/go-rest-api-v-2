package container

import (
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/order"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/user"
)

// Repositories holds all data access layer implementations.
type Repositories struct {
	User  user.Repository
	Order order.Repository
}

func (c *Container) buildRepositories() *Repositories {
	return &Repositories{
		User:  user.NewRepository(c.DB),
		Order: order.NewRepository(c.DB),
	}
}
