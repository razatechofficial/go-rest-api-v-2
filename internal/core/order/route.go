package order

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	orders := rg.Group("/orders")
	{
		orders.POST("", h.Create)
		orders.GET("", h.List)
		orders.GET("/:id", h.GetByID)
		orders.PUT("/:id", h.Update)
		orders.DELETE("/:id", h.Delete)
	}
}
