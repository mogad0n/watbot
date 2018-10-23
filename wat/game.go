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
	me Player
	commands map[string](func(*Player,[]string)(string))
	lifeCommands map[string](func(*Player, []string)(string))
}

func NewWatGame(bot *WatBot, db *WatDb) *WatGame {
	g := WatGame{bot, db, Player{}, nil, nil}
	g.me = g.db.User(bot.Nick, "tripsit/user/"+bot.Nick, true)
	g.commands = map[string](func(*Player,[]string)(string)) {
		"wat": g.megaWat,
		"watch": g.Watch,
		"watcoin": g.Balance,
		"send": g.Send,
		"rest": g.Rest,
		"leech": g.Leech,
		"roll": g.Roll,
		"dice": g.Dice,
	}
	g.lifeCommands = map[string](func(*Player, []string)(string)) {
		"steal": g.Steal,
		"frame": g.Frame,
		"punch": g.Punch,
		"quest": g.QuestStart,
	}
	return &g
}

var currency = "watcoin"
var currencys = "watcoins"
var unconscious = "wat, your hands fumble and fail you. try resting, weakling."
var helpText = fmt.Sprintf("watcoin <nick>, watch <nick>, topten, mine, send <nick> <%s>, roll <%s>, steal <nick> <%s>, frame <nick> <%s>, punch <nick>", currency, currency, currency, currency)
var rules = "A new account is created with 5 hours time credit. Mining exchanges time credit for %s: 1-10h: 1 p/h; >10h: 10 p/h; >1 day: 50 p/h; >1 month: 1000 p/h."

// missing
// invent, create, give inventory

func (g *WatGame) Msg(m *irc.Message, player *Player, fields []string) {
	reply := ""
	if g.commands[fields[0]] != nil {
		reply = g.commands[fields[0]](player, fields)
	} else {
		switch strings.ToLower(fields[0]) {
		case "rules":
			reply = rules
		case "help":
			reply = helpText
		case "topten":
			reply = fmt.Sprintf("%s holders: %s", currency, g.TopTen())
		case "mine":
			reply = g.Mine(player)
		}
	}
	if g.lifeCommands[fields[0]] != nil {
		// Nothing was handled. Maybe this is an action that requires consciousness.
		if !player.Conscious() {
			reply = unconscious
		} else {
			reply = g.lifeCommands[fields[0]](player, fields)
		}
	}
	g.bot.reply(m, reply)
}

func (g *WatGame) Dice(player *Player, fields []string) string {
	roll := int64(6)
	if len(fields) > 1 {
		i, e := g.Int(fields[1])
		if e == nil {
			roll = i
		}
	}
	answer := rand.Int63n(roll)+1
	return fmt.Sprintf("1d%d - %d", roll, answer)
}

type PositiveError struct {}
func (e PositiveError) Error() string {return ""}

func (g *WatGame) Int(str string) (int64, error) {
	i, e := strconv.ParseInt(str, 10, 64)
	if i < 0 {
		return 0, PositiveError{}
	}
	return i, e
}

func (g *WatGame) Roll(player *Player, fields []string) string {
	if len(fields) < 2 {
		return fmt.Sprintf("roll <%s> pls", currency)
	}
	amount, e := g.Int(fields[1])
	if e != nil {
		return "wat kinda numba is that"
	}
	if amount > player.Coins {
		return "wat? brokeass"
	}
	n := rand.Int63n(100)+1
	ret := fmt.Sprintf("%s rolls a d100 (<50 wins): It's a %d! ", player.Nick, n)
	if n < 50 {
		player.Coins += amount
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
		dmg += player.Level(player.Anarchy)
		ret += fmt.Sprintf("hits %s right in the torso for %d points of damage!", target.Nick, dmg)
		target.Health -= dmg
		g.db.Update(target)
	} else {
		dmg += target.Anarchy
		ret += fmt.Sprintf("fumbles, and punches themselves in confusion! %d self-damage! ", dmg)
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
		return fmt.Sprintf("frame <nick> <%s> - d6 roll. Sneaky? You force the target to pay me. Clumsy? You pay a fine to the target and myself.", currency)
	}
	amount, e := g.Int(fields[2])
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
		ret += fmt.Sprintf("You frame %s for a minor crime. They pay me %d.", target.Nick, amount)
		player.Anarchy += 1
		target.Coins -= amount
		// bot gets coins
	} else {
		ret += fmt.Sprintf("You were caught and pay them %d. %s gets the rest.", (amount/2), g.bot.Nick)
		player.Coins -= amount
		target.Coins += amount/2
		g.me.Coins += amount/2
		g.db.Update(g.me)
	}
	g.db.Update(player)
	g.db.Update(target)
	return ret
}

func (g *WatGame) Steal(player *Player, fields []string) string {
	if len(fields) < 3 {
		return fmt.Sprintf("steal <nick> <%s> - d6 roll. If you fail, you pay double the %s to %s", currency, currency, g.bot.Nick)
	}
	amount, e := g.Int(fields[2])
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
		return fmt.Sprintf("wat? %s is a poor fuck and doesn't have that much to steal.", target.Nick)
	}
	n := rand.Int63n(6)+1
	ret := fmt.Sprintf("%s is trying to steal %d %s from %s... ", player.Nick, amount, currency, target.Nick)
	if n < 3 {
		ret += "! You snuck it! Sneaky bastard!"
		player.Coins += amount
		player.Anarchy += 1
		target.Coins -= amount
		g.db.Update(target)
	} else {
		ret += fmt.Sprintf("... You were caught and I took %d %s from your pocket.", (amount*2), currency)
		player.Coins -= amount*2
	}
	g.db.Update(player)
	return ret
}

func (g *WatGame) GetTarget(player, target string) (*Player, string) {
	t := g.db.User(strings.ToLower(target), "", false)
	if t.Nick == "" {
		return nil, "Who? wat?"
	}
	if t.Nick == player {
		return nil, "You can't do that to yourself, dummy."
	}
	return &t, ""
}

func (g *WatGame) Leech(player *Player, fields []string) string {
	divisor := int64(10)
	if len(fields) < 3 {
		return fmt.Sprintf("leech <nick> <%s> - using your wealth, you steal the life force of another player. this will probably backfire and kill you.", currency)
	}
	amount, er := g.Int(fields[2])
	if amount < divisor {
		return fmt.Sprintf("wat?? can't try to leech with less than %d %s, what do you think this is, free?", divisor, currency)
	}
	if player.Coins < amount || er != nil {
		return "wat great fortune do you think you have? poor wats shouldn't be doing this, wat a waste..."
	}
	target, err := g.GetTarget(player.Nick, fields[1])
	if err != "" {
		return err
	}
	r := rand.Int63n(10)+1
	reply := fmt.Sprintf("You muster your wealth and feed it to %s. ", g.bot.Nick)
	hpDown := amount/divisor
	player.Coins -= amount
	if r < 5 {
		target.Health -= hpDown
		player.Health += hpDown
		reply += fmt.Sprintf("The deal is done, you schlorped %d HP from %s. They now have %d HP, you have %d.", hpDown, target.Nick, target.Health, player.Health)
	} else {
		reply += "The gods do not smile upon you this waturday. Your money vanishes and nothing happens."
	}
	return reply
}

func (g *WatGame) Rest(player *Player, fields []string) string {
	minRest := int64(86400)
	delta := time.Now().Unix() - player.LastRested
	ret := ""
	if player.LastRested == 0 {
		ret = "you've never slept before - you sleep so well, your injuries are cured and your health is restored to 10"
		player.Health = 10
	} else if delta < minRest {
		ret = fmt.Sprintf("wat were you thinking, sleeping at a time like this (%d until next rest)", minRest-delta)
	} else {
		value := rand.Int63n(20)+1
		ret = fmt.Sprintf("wat a nap - have back a random amount of hitpoints (this time it's %d)", value)
	}
	player.LastRested = time.Now().Unix()
	g.db.Update(player)
	return ret
}

func (g *WatGame) QuestStart(player *Player, fields []string) string {
	// Begin a quest with some people. It will involve multiple dice rolls.
	return ""
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
		player.Coins += value
		msg = fmt.Sprintf("%s mined %d %s for %d and has %d %s", player.Nick, value, currency, delta, player.Coins, currency)
	}
	player.LastMined = time.Now().Unix()
	g.db.Update(player)
	return msg
}

func (g *WatGame) Watch(player *Player, fields []string) string {
	if len(fields) > 1 {
		maybePlayer, err := g.GetTarget("", fields[1])
		if err != "" {
			return err
		}
		player = maybePlayer
	}
	return fmt.Sprintf("%s's Watting: %d (%d) / Anarchy: %d (%d) / Trickery: %d (%d) / Coins: %d / Health: %d", player.Nick, player.Level(player.Watting), player.Watting, player.Level(player.Anarchy), player.Anarchy, player.Trickery, player.Trickery, player.Coins, player.Health)
}

func (g *WatGame) Balance(player *Player, fields []string) string {
	balStr := "%s's %s balance: %d. Mining time credit: %d."
		balPlayer := player
		if len(fields) > 1 {
			var err string
			balPlayer, err = g.GetTarget("", fields[1])
			if err != "" {
				return err
			}
		}
		return fmt.Sprintf(balStr, balPlayer.Nick, currency, balPlayer.Coins, time.Now().Unix()-balPlayer.LastMined)
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

func (g *WatGame) megaWat(player *Player, _ []string) string {
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
