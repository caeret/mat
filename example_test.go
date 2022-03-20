package mat_test

import (
	"log"
	"net/http"

	"github.com/caeret/mat/access"
	"github.com/caeret/mat/content"
	"github.com/caeret/mat/fault"
	"github.com/caeret/mat/file"
	"github.com/caeret/mat/slash"
)

func Example() {
	router := mat.New()

	router.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	// serve RESTful APIs
	api := router.Group("/api")
	api.Use(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON, content.XML),
	)
	api.Get("/users", func(c *mat.Context) error {
		return c.Write("user list")
	})
	api.Post("/users", func(c *mat.Context) error {
		return c.Write("create a new user")
	})
	api.Put(`/users/<id:\d+>`, func(c *mat.Context) error {
		return c.Write("update user " + c.Param("id"))
	})

	// serve index file
	router.Get("/", file.Content("ui/index.html"))
	// serve files under the "ui" subdirectory
	router.Get("/*", file.Server(file.PathMap{
		"/": "/ui/",
	}))

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
