package dbc

import (
	"log"

	"github.com/jinzhu/gorm"

	// import sqlite dialect
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// DB is the gorm db instance
var DB *gorm.DB

// User struct is our user model for sqlitedb
type User struct {
	gorm.Model
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Token     string `json:"token"`
}

// Open opens sqlite db
func Open() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// set var db
	DB = db

	// Migrate the schema
	db.AutoMigrate(&User{})

	// return db instance
	return db
}

// Save saves user
func (u *User) Save() {
	DB.Create(u)
}
