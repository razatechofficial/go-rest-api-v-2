// internal/container/handlers.go
package container

// future: "myapp/internal/core/order"

// Handlers holds all HTTP delivery layer implementations.
// Handlers receive service interfaces — never concrete service types.
type Handlers struct {
	// User *user.Handler
	// Order   *order.Handler
	// Product *product.Handler
}

func (c *Container) buildHandlers() *Handlers {
	return &Handlers{
		// User: user.NewHandler(c.Services.User),
		// Order:   order.NewHandler(c.Services.Order),
		// Product: product.NewHandler(c.Services.Product),
	}
}
