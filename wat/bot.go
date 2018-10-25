package wat

import (
	"github.com/go-irc/irc"
	"crypto/tls"
	"fmt"
	"strings"
)

type WatBot struct {
	client *irc.Client
	conn *tls.Conn
	game *WatGame
	Db *WatDb
	Nick string
}

var allowedChannels = []string {
	"##wat",
	"##test",
	"##sweden",
}

func NewWatBot(config *irc.ClientConfig, serverConn *tls.Conn) *WatBot {
	wat := WatBot{conn:serverConn, Nick:config.Nick}
	wat.Db = NewWatDb()
	wat.game = NewWatGame(&wat, wat.Db)
	config.Handler = irc.HandlerFunc(wat.HandleIrcMsg)
	wat.client = irc.NewClient(wat.conn, *config)
	return &wat
}

func CleanNick(nick string) string {
	return string(nick[0])+"\u200c"+nick[1:]
}

func (w *WatBot) HandleIrcMsg(c *irc.Client, m *irc.Message) {
	switch cmd := m.Command; cmd {
	case "PING":
		w.write("PONG", m.Params[0])
	case "PRIVMSG":
		w.Msg(m)
	}
}

func (w *WatBot) Admin(m *irc.Message) bool {
	return m.Prefix.Host == "tripsit/operator/hibs"
}

func (w *WatBot) AllowedChannel(c string) bool {
	for _, allowed := range allowedChannels {
		if c == allowed {
			return true
		}
	}
	return false
}

func (w *WatBot) Msg(m *irc.Message) {
	// bail out if you're not yves, if you're not tripsit or if you're not in an allowed channel
	// but if you're an admin you can do whatever
	if m.Prefix.Host == "tripsit/user/Yves" || !strings.Contains(m.Prefix.Host, "tripsit") || (!w.AllowedChannel(m.Params[0]) && !w.Admin(m)) {
		return
	}

	// make sure there's actually some text to process
	if len(m.Params[1]) == 0 {
		return
	}

	// fieldsfunc allows you to obtain rune separated fields/args
	args := strings.FieldsFunc(m.Params[1], func(c rune) bool {return c == ' '})

	if len(args) == 0 {
		return
	}

	if w.Admin(m) {
		// allow impersonation of the robot from anywhere
		if args[0] == "imp" && len(args) > 2 {
			if args[1] == "PRIVMSG" {
				w.write(args[1], strings.Join(args[2:], " "))
			} else {
				w.write(args[1], args[2:]...)
			}
			return
		}
	}

	// strip offline message prefix from znc for handling offline buffer
	if args[0][0] == '[' && len(args) > 1 {
		args = args[1:]
	}

	// check if command char (or something weird) is present
	if args[0] != "wat" && args[0][0] != '#' {
		return
	}


	// clean input
	if args[0][0] == '#' {
		args[0] = args[0][1:]
	}

	user := strings.ToLower(m.Prefix.Name)
	player := w.Db.User(user, m.Prefix.Host, true)
	w.game.Msg(m, &player, args)
}

func (w *WatBot) Run() {
	defer w.conn.Close()
	err := w.client.Run()
	if err != nil {
		fmt.Println("Error returned while running client: " + err.Error())
	}
}

func (w *WatBot) say(dest, msg string) {
	if len(msg) == 0 {
		return
	}
	fmt.Printf("MSG %s: %s\n", dest, msg)
	w.write("PRIVMSG", dest, msg)
}

func (w *WatBot) reply(s *irc.Message, r string) {
	sender := s.Params[0]
	if sender == w.Nick {
		sender = s.Prefix.Name
	}
	w.say(sender, r)
}

func (w *WatBot) write(cmd string, params ...string) {
	w.client.WriteMessage(&irc.Message{
		Command: cmd,
		Params: params,
	})
}
