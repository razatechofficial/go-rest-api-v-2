// internal/container/repositories.go
package container

// future: "myapp/internal/core/order"
// future: "myapp/internal/core/product"

// Repositories holds all data access layer implementations.
// All fields typed as interfaces — never concrete types.
type Repositories struct {
	// User user.Repository
	// Order   order.Repository
	// Product product.Repository
}

func (c *Container) buildRepositories() *Repositories {
	return &Repositories{
		// User: user.NewRepository(c.db),
		// Order:   order.NewRepository(c.db),
		// Product: product.NewRepository(c.db),
	}
}
