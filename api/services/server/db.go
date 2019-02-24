package server

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
)

// ConnectToDb connects to mongodb and sets session
// so we can defer session close in main routine
func (s *Server) ConnectToDb() {
	var err error

	// establish connection with mongo
	s.Session, err = mgo.DialWithTimeout(os.Getenv("MONGODB_URI"), time.Second*3)
	if err != nil {
		log.Fatalf("Unable to connect to database: %s", err.Error())
	}

	// extract DB name from uri
	uriParts := strings.Split(os.Getenv("MONGODB_URI"), "/")
	dbName := uriParts[len(uriParts)-1]

	// set server db
	s.Db = s.Session.DB(dbName)
}
