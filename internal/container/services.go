// internal/container/services.go
package container

import "github.com/razatechofficial/go-rest-api-v-2/internal/core/user"

// future: "myapp/internal/core/order"

// Services holds all business logic layer implementations.
// Services receive repository interfaces — never concrete repo types.
type Services struct {
	User user.Service
	// Order   order.Service
	// Product product.Service
}

func (c *Container) buildServices() *Services {
	return &Services{
		User: user.NewService(c.Repositories.User),
		// Order: order.NewService(
		//     c.Repositories.Order,
		//     c.cfg,
		// ),
	}
}
