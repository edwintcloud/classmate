package server

import (
	"fmt"
	"net/http"
	"os"

	uuid "github.com/satori/go.uuid"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// NewError is type to reference our server Error function
// log.go
type NewError func(error, int) bson.M

// NewSuccess is type to reference our server Success function
// log.go
type NewSuccess func() bson.M

// Server is our echo server struct
type Server struct {
	Echo      *echo.Echo
	Db        *mgo.Database
	Log       *os.File
	Session   *mgo.Session
	JwtSecret []byte
}

// EchoHandler registers echo controllers with echo
func EchoHandler() *Server {
	server := Server{}

	// initilaize logger/error handling
	server.InitLogger("server_logs.txt")

	// connect to db
	server.ConnectToDb()

	// set JwtSecret to random uuid
	server.JwtSecret = []byte(uuid.NewV4().String())

	// create new instance of echo web serer
	server.Echo = echo.New()

	// register logging middleware with echo instance
	server.Echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method}  ${uri}  ${latency_human}  ${status}\n",
	}))

	// catch all route
	server.Echo.Any("*", func(c echo.Context) error {
		err := fmt.Sprintf("Bad Request - %s %s", c.Request().Method, c.Request().RequestURI)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err})
	})

	// return server instance
	return &server
}
