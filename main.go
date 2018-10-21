package main

import "fmt"
import "github.com/go-irc/irc"
import "github.com/namsral/flag"
import "crypto/tls"

import "git.circuitco.de/self/watbot/wat"

func main() {
	pass := flag.String("pass", "", "password")
	flag.Parse()
	config := irc.ClientConfig {
		Nick: "watt",
		Pass: *pass,
		User: "wat/tripsit",
		Name: "wat",
	}
	tcpConf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:9696", tcpConf)
	if err != nil {
		fmt.Println("err " + err.Error())
		return
	}
	wwat := wat.NewWatBot(&config, conn)
	wwat.Run()
}
