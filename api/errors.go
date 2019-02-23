package main

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
func Error(msg string, status int) bson.M {
	log.Println(msg)
	return bson.M{
		"error":  msg,
		"status": status,
	}
}
