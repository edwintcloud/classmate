package main

import (
	"os"

	"github.com/edwintcloud/classmate/api/services/attendance"
	"github.com/edwintcloud/classmate/api/services/server"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	server := server.EchoHandler()

	// defer log file close
	defer server.Log.Close()

	// defer mgo session close
	defer server.Session.Close()

	// register other services with server
	attendance.Register(server)

	// start http server
	server.Echo.Logger.Fatal(server.Echo.Start(":" + os.Getenv("PORT")))
}
