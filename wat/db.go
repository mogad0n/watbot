package wat

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Player struct {
	gorm.Model
	Nick       string
	Host       string
	Watting    int64
	Anarchy    int64
	Trickery   int64
	Coins      uint64 `gorm:"default:'100'"`
	Health     int64
	LastMined  int64
	LastRested int64
	CoinsLost  uint64
}

type Action struct {
	PlayerId  uint       `gorm:"primary_key;auto_increment:false"`
	Type      ActionType `gorm:"primary_key;auto_increment:false"`
	Performed int64
}

func (p *Player) LoseCoins(coins uint64) {
	p.Coins -= coins
	p.CoinsLost += coins
}

func (p *Player) Conscious() bool {
	return (p.Health > 0)
}

func (p *Player) Level(xp int64) int64 {
	if xp < 100 {
		return xp / 10
	} else if xp < 900 {
		return 10 + (xp / 100)
	} else {
		return 99
	}
}

type WatDb struct {
	db *gorm.DB
}

func NewWatDb() *WatDb {
	w := WatDb{}
	var err error
	w.db, err = gorm.Open("sqlite3", "wat.db")
	if err != nil {
		panic(err)
	}
	w.db.AutoMigrate(&Action{}, &Player{})
	return &w
}

func (w *WatDb) User(nick, host string, create bool) Player {
	var player Player
	// Try and get a user
	if err := w.db.First(&player, "nick = ? or host = ?", nick, host).Error; err != nil && create {
		fmt.Printf("Creating user: %s\n", err.Error())
		// No user, make another
		player.Nick = nick
		player.Host = host
		w.db.Create(&player)
		w.db.First(&player, "nick = ? or host = ?", nick, host)
	}
	return player
}

func (w *WatDb) Update(upd ...interface{}) {
	for _, u := range upd {
		fmt.Printf("Updating %+v\n", u)
		w.db.Save(u)
	}
}

const (
	Action_Mine ActionType = 1
	Action_Rest ActionType = 2
	Action_Lift ActionType = 3
	Action_Riot ActionType = 4
)

type ActionType int

func (w *WatDb) LastActed(player *Player, actionType ActionType) int64 {
	action := Action{}
	w.db.First(&action, "type = ? AND player_id = ?", actionType, player.Model.ID)
	return action.Performed
}

func (w *WatDb) Act(player *Player, actionType ActionType) {
	action := Action{player.Model.ID, actionType, time.Now().Unix()}
	if w.db.First(&action, "type = ? AND player_id = ?", actionType, player.ID).RecordNotFound() {
		w.db.Create(&action)
	} else {
		action.Performed = time.Now().Unix()
		w.Update(&action)
	}
}

func (w *WatDb) TopLost() []Player {
	var user = make([]Player, 10)
	w.db.Limit(10).Order("coins_lost desc").Find(&user)
	return user
}

func (w *WatDb) TopTen() []Player {
	var user = make([]Player, 10)
	w.db.Where("nick != 'watt'").Limit(10).Order("coins desc").Find(&user)
	return user
}
