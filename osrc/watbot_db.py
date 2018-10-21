#!/usr/bin/python

import sqlite3
import time


class WatbotDB:
    """watbot database abstraction"""
    
    def __init__(self, config):
      self.config = config
      self.currency = config["currency"]
      self.conn = sqlite3.connect(self.currency + ".db")
      self.c = self.conn.cursor()

      self.c.execute("create table if not exists players ("
                     "  nickname   varchar(256),"
                     "  watting    integer,"
                     "  anarchy    integer, "
                     "  trickery   integer, "
                     "  coins      integer, "
                     "  health     integer, "
                     "  last_mine  integer, "
                     "  last_rest  integer, "
                     "  primary key(nickname)"
                     ")"
      )

      self.c.execute("create table if not exists ledger ("
                     "  nickname   varchar(256),"
                     "  timestamp  integer,"
                     "  balance    integer,"
                     "  log        text,"
                     "  foreign key(nickname) references players(nickname)"
                     ")"
      )

      self.c.execute("create table if not exists items ("
                     "  itemname   varchar(256),"
                     "  nickname   varchar(256),"
                     "  price integer,"
                     "  primary key(itemname),"
                     "  foreign key(nickname) references players(nickname)"
                     ")"
      )
                
      self.c.execute("create table if not exists inventory ("
                     "  nickname  varchar(256),"
                     "  itemname  varchar(256),"
                     "  count     integer,"
                     "  foreign key(nickname) references players(nickname),"
                     "  foreign key(itemname) references items(itemname)"
                     ")"
      )

      self.conn.commit()
      
    def get_account(self,nick):
       self.c.execute("select watting, anarchy, trickery, coins, last_mine, health, last_rest from players where nickname=?", (nick.lower(),))
       data = self.c.fetchone()
       
       if data is None:
          earlier = time.time() - ( 24 * 3600 + 1)
          self.c.execute("insert into players(nickname, watting, anarchy, trickery, coins, last_mine, health, last_rest) values(?, 0, 0, 0, 0, ?, 100, ?)", (nick.lower(), earlier, earlier))
          self.c.execute("insert into ledger(nickname, timestamp, balance, log) values(?, ?, ?, ?)", (nick.lower(), earlier, 0, "created"))
          self.conn.commit()
          data = (0, 0, 0, 0, earlier, 100, earlier)
       
       return data

    def update_account(self, nick, watting, anarchy, trickery, coins, last_mine, health, last_rest, log):
       now = time.time()
       self.c.execute("update players set watting=?, anarchy=?, trickery=?, coins=?, last_mine=?, health=?, last_rest=? where nickname=?", (watting, anarchy, trickery, coins, last_mine, health, last_rest, nick.lower()))
       self.c.execute("insert into ledger(nickname, timestamp, balance, log) values(?, ?, ?, ?)", (nick.lower(), now, coins, log))
       if not log is None:
           print "log: " + log
       self.conn.commit()

    def close(self):
       self.conn.close()

    def topten(self):
        out = ""
        self.c.execute("select nickname, coins from players order by coins desc limit 10")
        while True:
           d = self.c.fetchone()
           if d == None:
              break
           out += d[0] + "(" + str(d[1]) + ") "
           
        return out   

    def items(self):
        out = ""
        self.c.execute("select itemname, nickname, price from items order by itemname")
        while True:
           d = self.c.fetchone()
           if d == None:
              break
           out += d[0] + "(" + d[1] + ", " + str(d[2]) + ") "
           
        return out   

    def invent_item(self, itemname, nickname, price):
        try:
            self.c.execute("insert into items(itemname, nickname, price) values(?, ?, ?)", (itemname.lower(), nickname.lower(), price))
            self.conn.commit()
            return True
        except:
            return False
          
    def inventory(self, nickname):
        out = ""
        self.c.execute("select itemname, count from inventory where nickname = ? order by itemname", (nickname.lower(),))
        while True:
            d = self.c.fetchone()
            if d == None:
                break
            out += d[0] + "(" + str(d[1]) + ") "
           
        return out

    def get_item(self, itemname):
        self.c.execute("select itemname, nickname, price from items where itemname = ?", (itemname.lower(),))
        return self.c.fetchone()

    def get_item_count(self, nickname, itemname):
        self.c.execute("select count from inventory where nickname = ? and itemname = ?", (nickname.lower(), itemname.lower()))
        d = self.c.fetchone()
        if d == None:
            return 0
        else:
            return d[0]
                       
    def set_item_count(self, nickname, itemname, count):
        if count == 0:
            self.c.execute("delete from inventory where nickname = ? and itemname = ?", (nickname.lower(), itemname.lower()))
        else:
            if self.get_item_count(nickname, itemname) != 0:
                self.c.execute("update inventory set count=? where nickname = ? and itemname = ?", (count, nickname.lower(), itemname.lower()))
            else:
                self.c.execute("insert into inventory(nickname, itemname, count) values(?, ?, ?)", (nickname.lower(), itemname.lower(), count))
        self.conn.commit()
