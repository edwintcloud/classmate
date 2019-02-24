package server

import (
	"io"
	"log"
	"os"

	"github.com/globalsign/mgo/bson"
)

// InitLogger sets up logging for server errros
func (s *Server) InitLogger(name string) {
	var err error

	// open file
	s.Log, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Unable to initialize logger:", err.Error())
	}

	// setup logger to use file and stdout
	wrt := io.MultiWriter(os.Stdout, s.Log)
	log.SetOutput(wrt)
}

// Error handles errors for our server by logging to
// stdout and file
func Error(err error, status int) bson.M {
	log.Println(err.Error())
	return bson.M{
		"error":  err.Error(),
		"status": status,
	}
}

// Success handles success messages for our server
// by returning json
func Success() bson.M {
	return bson.M{
		"message": "OK",
		"status":  200,
	}
}
