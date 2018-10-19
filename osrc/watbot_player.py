#!/usr/bin/python

import time

class WatbotPlayer:
    """class representing a player account"""
    
    def __init__(self, db, nick):
        self.nick = nick
        self.db = db
        
        ( 
          self.watting_exp, 
          self.anarchy_exp, 
          self.trickery_exp, 
          self.coins,
          self.last_mine,
          self.health,
          self.last_rest
        ) = db.get_account(nick)
        
        self.watting = self.get_level(self.watting_exp)
        self.anarchy = self.get_level(self.anarchy_exp)
        self.trickery = self.get_level(self.trickery_exp)
        
        now = time.time()
        delta = now - self.last_rest
        if delta > 60:
            self.health += int(delta/60)
            if self.health > 10:
              self.health = 10
              
            self.last_rest += int(delta/60) * 60

    def get_level(self, exp):
        if exp < 100:
            level = int(exp/10)
        elif exp < 900:
            level = 10 + int(exp/100)
        else:
            level = 99
            
        return level

    def update(self, log):
        self.db.update_account(self.nick, self.watting_exp, self.anarchy_exp, self.trickery_exp, self.coins, self.last_mine, self.health, self.last_rest, log)
