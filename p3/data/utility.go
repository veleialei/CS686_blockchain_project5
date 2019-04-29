package data

import "fmt"

func PrintError(err error, function string) {
	fmt.Println("The error function " + function + "cause the error: " + err.Error())
}
