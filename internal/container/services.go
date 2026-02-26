package container

import (
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/order"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/user"
)

// Services holds all business logic layer implementations.
type Services struct {
	User  user.Service
	Order order.Service
}

func (c *Container) buildServices() *Services {
	return &Services{
		User:  user.NewService(c.Repositories.User),
		Order: order.NewService(c.Repositories.Order),
	}
}
