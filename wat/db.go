package wat

import (
	"time"
	"github.com/jinzhu/gorm"
	_"github.com/jinzhu/gorm/dialects/sqlite"
)
import "fmt"

type Player struct {
	gorm.Model
	Nick string
	Host string
	Watting int64
	Anarchy int64
	Trickery int64
	Coins int64 `gorm:"default:'100'"`
	Health int64
	LastMined int64
	LastRested int64
}

func (p *Player) Conscious() bool {
	return (p.Health > 0)
}

func (p *Player) Level(xp int64) int64 {
	if xp < 100 {
		return xp/10
	} else if xp < 900 {
		return 10 + (xp/100)
	} else {
		return 99
	}
}

type Ledger struct {
	PlayerId uint `gorm:"primary_key"`
	Time int64
	Balance int64
	Log string
}

type Item struct {
	PlayerId uint
	Name string `gorm:"primary_key"`
	Price int64
}

type PlayerItem struct {
	PlayerId uint
	ItemId int
	Count int
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
	w.db.AutoMigrate(&Player{}, &Ledger{}, &Item{}, &PlayerItem{})
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
		w.db.Create(&Ledger{player.Model.ID, time.Now().Unix(), 0, "creation"})
	}
	return player
}

func (w *WatDb) Update(upd interface{}) {
	w.db.Save(upd)
}

func (w *WatDb) TopTen() []Player {
	var user = make([]Player, 10)
	w.db.Limit(10).Order("coins desc").Find(&user)
	return user
}
