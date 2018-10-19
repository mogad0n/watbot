package main

import "fmt"
import "github.com/go-irc/irc"
import "github.com/namsral/flag"

func main() {
	var pass string
	flag.String(pass)
	config := irc.ClientConfig {
		Nick: "wat",
		Pass: pass,
		User: "wat",
		Name: "wat",
		Handler: testHandler
	}
	fmt.Println("Hello world")
}
