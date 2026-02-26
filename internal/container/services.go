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
	// 1. User service first, without OrderLister (avoids circular dependency)
	userSvc := user.NewService(c.Repositories.User, nil)
	// 2. Adapter: order module uses user service as ports.UserChecker
	userChecker := userCheckerAdapter{Service: userSvc}
	// 3. Order service with UserChecker so Create can reject inactive users
	orderSvc := order.NewService(c.Repositories.Order, userChecker)
	// 4. Adapter: user module uses order service as ports.OrderLister
	orderLister := orderListerAdapter{Service: orderSvc}
	// 5. Two-phase wiring: inject OrderLister into user service
	if setter, ok := userSvc.(user.OrderListerSetter); ok {
		setter.SetOrderLister(orderLister)
	}
	return &Services{
		User:  userSvc,
		Order: orderSvc,
	}
}
