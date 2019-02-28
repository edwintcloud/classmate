package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/edwintcloud/classmate/cli/dbc"
	"github.com/globalsign/mgo/bson"
	format "github.com/labstack/gommon/color"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	print     = format.Println
	underline = format.U
	bold      = format.B
	invert    = format.In
	dim       = format.D
	stop      = false
	user      *dbc.User
)

func main() {

	// open sqlite db
	db := dbc.Open()
	defer db.Close()

	start()

	print(format.Yellow("\nGoodbye!"))
}

func start() {
	user = dbc.GetUser()
	if user.Role == "student" {
		for !stop {
			studentDashboard()
		}
	} else if user.Role == "teacher" {
		for !stop {
			teacherDashboard()
		}
	} else if user.Role == "admin" {
		for !stop {
			adminDashboard()
		}
	} else {
		for !stop {
			dashboard()
		}
	}
}

// User Dashboard
func studentDashboard() {
	print(format.Underline("\nWelcome to Classmate "+user.FirstName+"!"), "\n")
	print(format.Green("1.) Check in to class"))
	print(format.Cyan("2.) List classes"))
	print(format.Red("3.) Exit"), "\n")
	print("Please make a selection:")
	choice := getInput()
	switch choice {
	case "1":
		signup()
	case "2":
		login()
	case "3":
		stop = true
	default:
		print(format.Magenta("\nInvalid input!"))
	}
}

// Teacher Dashboard
func teacherDashboard() {
	print(format.Underline("\nWelcome to Classmate "+user.FirstName+"!"), "\n")
	print(format.Green("1.) View current attendance"))
	print(format.Cyan("2.) List classes"))
	print(format.Red("3.) Exit"), "\n")
	print("Please make a selection:")
	choice := getInput()
	switch choice {
	case "1":
		signup()
	case "2":
		login()
	case "3":
		stop = true
	default:
		print(format.Magenta("\nInvalid input!"))
	}
}

// Admin Dashboard
func adminDashboard() {
	print(format.Underline("\nWelcome to Classmate "+user.FirstName+"!"), "\n")
	print(format.Green("1.) Add Class"))
	print(format.Cyan("2.) Enroll students"))
	print(format.Cyan("3.) List classes"))
	print(format.Cyan("4.) List users"))
	print(format.Red("5.) Exit"), "\n")
	print("Please make a selection:")
	choice := getInput()
	switch choice {
	case "1":
		signup()
	case "2":
		login()
	case "5":
		stop = true
	default:
		print(format.Magenta("\nInvalid input!"))
	}
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
		login()
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
	person["password"] = getHiddenInput()
	print(format.Cyan("Please confirm password:"))
	confirmPassword := getHiddenInput()
	print(format.Cyan("Please enter first name:"))
	person["first_name"] = getInput()
	print(format.Cyan("Please enter last name:"))
	person["last_name"] = getInput()

	// verify input
	if confirmPassword != person["password"] {
		print(format.Red("\nPasswords must match!"))
	} else {

		// create person
		err := dbc.CreatePerson(&person)
		if err != nil {
			print(format.Red("\n" + err.Error()))
		} else {
			start()
		}
	}
}

// Login
func login() {
	person := bson.M{}

	// get input
	print(format.Underline("\nLogin\n"))
	print(format.Cyan("Please enter email:"))
	person["email"] = getInput()
	print(format.Cyan("Please enter password:"))
	person["password"] = getHiddenInput()
	err := dbc.LoginPerson(&person)
	if err != nil {
		print(format.Red("\n" + err.Error()))
	} else {
		start()
	}

}

// Wait for input
func getInput() string {
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, _ := buf.ReadString('\n')
	return strings.Replace(sentence, "\n", "", -1)
}

// wait for input (hidden)
func getHiddenInput() string {
	fmt.Print("> ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	return string(bytePassword)
}
