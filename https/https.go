package https

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"streaming-project/database"
)

type Router struct {
	routerAddress  string
	routerEngine   *gin.Engine
	routerDatabase *database.Database
}

func NewRouter(address string, db *database.Database) *Router {
	router := gin.Default()

	newRouter := &Router{
		routerAddress:  address,
		routerEngine:   router,
		routerDatabase: db,
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers, access-control-allow-methods"},
	}))

	router.GET("/orders/:id", newRouter.getOrder)
	return newRouter
}

func (r *Router) Start() error {
	err := r.routerEngine.Run(r.routerAddress)
	if err != nil {
		return err
	}
	return nil
}

func (r *Router) getOrder(c *gin.Context) {
	orderId := c.Param("id")

	order, err := r.routerDatabase.GetOrderByID(orderId)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, "Order with this ID could not be found.")
		return
	}

	c.IndentedJSON(http.StatusOK, order)
}
