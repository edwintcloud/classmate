package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// NewError is type to reference our server Error function
// errors.go
type NewError func(msg string, status int) bson.M

// Server is our echo server struct
type Server struct {
	Echo    *echo.Echo
	Db      *mgo.Database
	Error   NewError
	Log     *os.File
	Session *mgo.Session
}

func main() {
	server := EchoHandler()

	// defer log file close
	defer server.Log.Close()

	// defer mgo session close
	defer server.Session.Close()

	// start http server
	server.Echo.Logger.Fatal(server.Echo.Start(":" + os.Getenv("PORT")))
}

// EchoHandler registers echo controllers with echo
func EchoHandler() *Server {
	server := Server{}

	// initilaize logger/error handling
	server.InitLogger("server_logs.txt")

	// connect to db
	server.ConnectToDb()

	// create new instance of echo web serer
	e := echo.New()

	// register logging middleware with echo instance
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method}  ${uri}  ${latency_human}  ${status}\n",
	}))

	// register services with api

	// catch all route
	e.Any("*", func(c echo.Context) error {
		err := fmt.Sprintf("Bad Request - %s %s", c.Request().Method, c.Request().RequestURI)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err})
	})

	// return server instance
	return &server
}
