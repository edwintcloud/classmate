package dbc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"
)

// This file will contain methods to connect to the api and make requests

var (
	client = &http.Client{
		Timeout: time.Second * 10,
	}
	host = "http://localhost:9000/api/v1"
)

// CreatePerson makes a post request to create a new person
func CreatePerson(person *bson.M) error {
	result := User{}

	// marshal data into json string
	bodyBytes, err := json.Marshal(person)
	if err != nil {
		return err
	}

	// make post request
	resp, err := client.Post(host+"/persons", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	// unmarshal result into person and store in localdb
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	// save user to localdb
	result.Save()

	return err
}

// LoginPerson makes a post request to login a person
func LoginPerson(person *bson.M) error {
	result := User{}

	// marshal data into json string
	bodyBytes, err := json.Marshal(person)
	if err != nil {
		return err
	}

	// make post request
	resp, err := client.Post(host+"/persons/login", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	// unmarshal result into result and store in localdb
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	// save user to localdb
	result.Save()

	return err
}
