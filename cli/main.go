package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	if choice == "3" {
		stop = true
	} else {
		print(format.Magenta("\nInvalid input!"))
	}
}

// Signup
func signup() {

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
