package wat

import (
	"fmt"
	"time"
	"strings"
	"strconv"
	"math/rand"
	"github.com/go-irc/irc"
)

type WatGame struct {
	bot *WatBot
	db *WatDb
}

func NewWatGame(bot *WatBot, db *WatDb) *WatGame {
	return &WatGame{bot, db}
}

var currency = "watcoin"
var currencys = "watcoins"
var unconscious = "wat, you're too weak for that."
var helpText = fmt.Sprintf("balance <nick>, watch <nick>, inventory <nick>, topten, mine, send <nick> <%s>, roll <%s>, steal <nick> <%s>, frame <nick> <%s>, punch <nick>", currency, currency, currency, currency)
var rules = "A new account is created with 5 hours time credit. Mining exchanges time credit for %s: 1-10h: 1 p/h; >10h: 10 p/h; >1 day: 50 p/h; >1 month: 1000 p/h."
// missing
// invent, create, give inventory

func (g *WatGame) Msg(m *irc.Message, player *Player, fields []string) {
	reply := ""
	switch strings.ToLower(fields[0]) {
	case "wat":
		reply = g.megaWat(player)
	case "rules":
		reply = rules
	case "watch":
		reply = fmt.Sprintf("Watting: %d (%d) / Anarchy: %d (%d) / Trickery: %d (%d) / Coins %d Health: %d", player.Level(player.Watting), player.Watting, player.Level(player.Anarchy), player.Anarchy, player.Trickery, player.Trickery, player.Coins, player.Health)
	case "help":
		reply = helpText
	case "topten":
		reply = fmt.Sprintf("%s holders: %s", currency, g.TopTen())
	case "balance":
		reply = g.Balance(player, fields)
	case "mine":
		reply = g.Mine(player)
	case "send":
		reply = g.Send(player, fields)
	}
	if reply == "" {
		// Nothing was handled. Maybe this is an action that requires consciousness.
		if !player.Conscious() {
			reply = unconscious
		} else {
			reply = g.WokeMsg(m, player, fields)
		}
	}
	g.bot.reply(m, reply)
}

func (g *WatGame) WokeMsg(m *irc.Message, player *Player, fields []string) string {
	reply := ""
	switch strings.ToLower(fields[0]) {
	case "steal":
		reply = g.Steal(player, fields)
	case "frame":
		reply = g.Frame(player, fields)
	case "punch":
		reply = g.Punch(player, fields)
	case "roll":
		reply = g.Roll(player, fields)
	}
	return reply
}

func (g *WatGame) Roll(player *Player, fields []string) string {
	if len(fields) < 2 {
		return fmt.Sprintf("roll <%s> pls", currency)
	}
	amount, e := strconv.ParseInt(fields[1], 10, 64)
	if e != nil {
		return "wat kinda numba is that"
	}
	if amount > player.Coins {
		return "wat? brokeass"
	}
	n := rand.Int63n(100)+1
	ret := fmt.Sprintf("%s rolls a d100 (<50 wins): It's a %d! ", player.Nick, n)
	if n < 50 {
		player.Coins -= amount
		ret += fmt.Sprintf("You win! Your new balance is %d", player.Coins)
	} else {
		player.Coins -= amount
		ret += fmt.Sprintf("You lose! Your new balance is %d", player.Coins)
	}
	g.db.Update(player)
	return ret
}

func (g *WatGame) Punch(player *Player, fields []string) string {
	if len(fields) < 2 {
		return "punch <target> pls"
	}
	target, err := g.GetTarget(player.Nick, fields[1])
	if err != "" {
		return err
	}
	chance := rand.Int63n(6)+1
	dmg := rand.Int63n(6)+1
	ret := fmt.Sprintf("%s rolls a d6 to punch %s: It's a %d! %s ", player.Nick, target.Nick, chance, player.Nick)
	if chance <3 {
		dmg += player.Anarchy
		ret += fmt.Sprintf("hits for %d points of damage!", dmg)
		target.Health -= dmg
		g.db.Update(target)
	} else {
		dmg += target.Anarchy
		ret += fmt.Sprintf("misses miserably! %s punches back for %d damage! ", target.Nick, dmg)
		player.Health -= dmg
		if player.Health <= 0 {
			ret += player.Nick + " has fallen unconscious."
		}
		g.db.Update(player)
	}
	return ret
}

func (g *WatGame) Frame(player *Player, fields []string) string {
	if len(fields) < 3 {
		return fmt.Sprintf("frame <nick> <%s> - d6 roll. If < 3, you force the target to pay me. Otherwise, you pay a fine to the target and myself.", currency)
	}
	amount, e := strconv.ParseInt(fields[2], 10, 64)
	if amount <= 0 || e != nil {
		return "wat kinda number is "+fields[2]+"?"
	}
	if player.Coins < amount {
		return "wat? you too poor for that."
	}
	target, err := g.GetTarget(player.Nick, fields[1])
	if err != "" {
		return err
	}
	if target.Coins < amount {
		return fmt.Sprintf("wat? %s is too poor for this.", target.Nick)
	}
	n := rand.Int63n(6)+1
	ret := fmt.Sprintf("%s rolls a d6 to frame %s with %d %s: It's a %d! (<3 wins). ", player.Nick, target.Nick, amount, currency, n)
	if n < 3 {
		ret += fmt.Sprintf("You win! They pay me %d.", amount)
		player.Anarchy += 1
		target.Coins -= amount
		// bot gets coins
	} else {
		ret += fmt.Sprintf("You were caught and pay them %d, throwing away the rest.", (amount/2))
		player.Coins -= amount
		target.Coins += amount/2
//		target.Coins += amount/2
	}
	g.db.Update(player)
	g.db.Update(target)
	return ret
}

func (g *WatGame) Steal(player *Player, fields []string) string {
	if len(fields) < 3 {
		return fmt.Sprintf("steal <nick> <%s> - d6 roll. If < 3, you steal the %s. Otherwise, you pay double the %s to %s", currency, currency, currency, g.bot.Nick)
	}
	amount, e := strconv.ParseInt(fields[2], 10, 64)
	if amount <= 0 || e != nil {
		return "wat kinda number is "+fields[2]+"?"
	}
	if player.Coins < amount*2 {
		return "wat? You'd go bankrupt if they steal back..."
	}
	target, err := g.GetTarget(player.Nick, fields[1])
	if target == nil {
		return err
	}
	if target.Coins < amount {
//		return fmt.Sprintf("wat? %s is a poor fuck and doesn't have that much to steal.", target.Nick)
	}
	n := rand.Int63n(6)+1
	ret := fmt.Sprintf("%s rolls a d6 to steal %d %s from %s... It's %d! (<3 wins) ", player.Nick, amount, currency, target.Nick, n)
	if n < 3 {
		ret += "You win! Sneaky bastard!"
		player.Coins += amount
		player.Anarchy += 1
		target.Coins -= amount
		g.db.Update(target)
	} else {
		ret += fmt.Sprintf("You were caught and I took %d %s from your pocket.", (amount*2), currency)
		player.Coins -= amount*2
	}
	g.db.Update(player)
	return ret
}

func (g *WatGame) GetTarget(player, target string) (*Player, string) {
	t := g.db.User(target, "", false)
	if t.Nick == "" {
		return nil, "Who?"
	}
	if t.Nick == player {
		return nil, "You can't do that to yourself, dummy."
	}
	return &t, ""
}

func (g *WatGame) Send(player *Player, fields []string) string {
	if len(fields) < 3 {
		return fmt.Sprintf("You forgot somethin'. send <nick> <%s>", currency)
	}
	amount, err := strconv.Atoi(fields[2])
	if err != nil {
		return fields[2] + " is not an integer, wat?"
	}
	if int64(amount) > player.Coins {
		return "wat? you poor fuck, you don't have enough!"
	}
	target, str := g.GetTarget(player.Nick, fields[1])
	if target == nil {
		return str
	}
	player.Coins -= int64(amount)
	target.Coins += int64(amount)
	g.db.Update(player)
	g.db.Update(target)
	return fmt.Sprintf("%s sent %s %d %s", player.Nick, target.Nick, amount, currency)
}

func (g *WatGame) Mine(player *Player) string {
	delta := time.Now().Unix() - player.LastMined
	if delta < 1800 {
				return fmt.Sprintf("wat? 2 soon (%d)", delta)
	}
	value := int64(0)
	if delta < 36000 {
		value = delta/1800
	} else if delta < 86400 {
		value = 10
	} else if delta < 2592000 {
		value = 50
	} else {
		value = 1000
	}
	msg := ""
	if player.LastMined == 0 {
		msg = "with wat? you go to get a pickaxe"
		value = 0
	} else {
		msg = fmt.Sprintf("%s mined %d %s for %d and has %d %s", player.Nick, value, currency, delta, player.Coins, currency)
	}
	player.Coins += value
	player.LastMined = time.Now().Unix()
	g.db.Update(player)
	return msg
}

func (g *WatGame) Balance(player *Player, fields []string) string {
	balStr := "%s's %s balance: %d. Mining time credit: %d."
		balPlayer := *player
		if len(fields) > 1 {
			balPlayer = g.db.User(fields[1], "", false)
			if balPlayer.Nick == "" {
				return "Who?"
			}
		}
		return fmt.Sprintf(balStr, balPlayer.Nick, currency, balPlayer.Coins, balPlayer.LastMined)
}

func (g *WatGame) TopTen() string {
	players := g.db.TopTen()
	ret := ""
	for _, p := range players {
		ret += PrintTwo(p.Nick, p.Coins)
	}
	return ret
}

func PrintTwo(nick string, value int64) string {
	return fmt.Sprintf("%s (%d) ", CleanNick(nick), value)
}

func (g *WatGame) megaWat(player *Player) string {
	mega := rand.Int63n(1000000)+1
	kilo := rand.Int63n(1000)+1
	reply := ""
	if mega == 23 {
		player.Coins += 1000000
		reply = fmt.Sprintf("OMGWATWATWAT!!!! %s has won the MegaWat lottery and gains 1000000 %s!", player.Nick, currency)
	}
	if kilo == 5 {
		player.Coins += 1000
		reply = fmt.Sprintf("OMGWAT! %s has won the KiloWat lottery and gains 1000 %s!", player.Nick, currency)
	}
	player.Watting += 1
	g.db.Update(player)
	return reply
}
