package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/edwintcloud/classmate/cli/dbc"
	"github.com/globalsign/mgo/bson"
	format "github.com/labstack/gommon/color"
)

var (
	print     = format.Println
	underline = format.U
	bold      = format.B
	invert    = format.In
	dim       = format.D
	stop      = false
)

func main() {
	for !stop {
		dashboard()
	}
	print(format.Yellow("\nGoodbye!"))
}

// Dashboard
func dashboard() {
	print(format.Underline("\nWelcome to Classmate!"), "\n")
	print(format.Green("1.) Create Account"))
	print(format.Cyan("2.) Login"))
	print(format.Red("3.) Exit"), "\n")
	print("Please make a selection:")
	choice := getInput()
	switch choice {
	case "1":
		signup()
	case "2":
	case "3":
		stop = true
	default:
		print(format.Magenta("\nInvalid input!"))
	}
}

// Signup
func signup() {
	person := bson.M{}

	// get input
	print(format.Underline("\nCreate an Account\n"))
	print(format.Cyan("Please enter email:"))
	person["email"] = getInput()
	print(format.Cyan("Please enter password:"))
	person["password"] = getInput()
	print(format.Cyan("Please confirm password:"))
	confirmPassword := getInput()
	print(format.Cyan("Please enter first name:"))
	person["first_name"] = getInput()
	print(format.Cyan("Please enter last name:"))
	person["last_name"] = getInput()

	// verify input
	if confirmPassword != person["password"] {
		print(format.Red("\nPasswords must match!"))
		signup()
	} else {

		// create person
		err := dbc.CreatePerson(&person)
		if err != nil {
			print(format.Red("\n" + err.Error()))
			signup()
		}
	}
}

// Login
func login() {

}

// Wait for input
func getInput() string {
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, _ := buf.ReadString('\n')
	return strings.Replace(sentence, "\n", "", -1)
}
