package main

import "fmt"
import "github.com/go-irc/irc"
import flag"github.com/namsral/flag"

func testHandler(c *irc.Client, m *irc.Message) {
	fmt.Println("Client: %+v", c)
}

func main() {
	pass := flag.String("pass", "", "password")
	flag.Parse()
	config := irc.ClientConfig {
		Nick: "wat",
		Pass: *pass,
		User: "wat",
		Name: "wat",
		Handler: irc.HandlerFunc(testHandler),
	}
	fmt.Printf("Hello world %+v\n", config)
}
