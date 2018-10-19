#!/usr/bin/python

import time
import random
import math

from watbot_player import WatbotPlayer

class WatbotGame:
    """Class containing the game logic"""
  
    def __init__(self, config, db):
        random.seed()

        self.config = config
        self.db = db
      
        self.bot_nick = config["bot_nick"]
        self.bot_player = WatbotPlayer(self.db, self.bot_nick)
        
        self.quiet = False

        self.help_text = (
         "balance [<nickname>], "
         "watch [<nickname>], "
         "inventory [<nickname>], "
         "topten, "
         "items, "
         "mine, " 
         "transfer <nickname> <amount>, "
         "roll <amount>, "
         "steal <nickname> <amount>, "
         "frame <nickname> <amount>, "
         "punch <nickname>, "
         "give <nickname> <count> <itemname>, "
         "invent <itemname> <price>, "
         "create <itemname> <count>, "
         
         )
         
        self.rules_text = ("A new account is created with 5 hours time credit. "
                    "Mining exchanges time credit for coins: "
                    "1h - 10h: 1 coin/h; >10h: 10 coin; >1 day: 50 coin; >1 month: 1000 coin.")
          
    def do_command(self, nick, command, args):
        try:
            player = WatbotPlayer(self.db, nick)
            self.now = time.time()

            if command == "wat":
                out = self.mega_wat(player)

            elif command == "speak" and (player.nick == "Burrito" or player.nick == "wuselfuzz"):
                self.quiet = False
                out = "wat!"

            elif command == "shutup" and (player.nick == "Burrito" or player.nick == "wuselfuzz"):
                self.quiet = True
                out = "wat."

            elif self.quiet:
                out = None

            elif command == "help":
                out = self.help_text

            elif command == "rules":
                out = self.rules_text

            elif command == "topten":
                out = self.db.topten()
                
            elif command == "items":
                out = "temporarily disabled, because bug!"
                #out = self.db.items()

            elif command == "inventory":
                if len(args) > 0:
                    out = self.db.inventory(args[0])
                else:
                    out = self.db.inventory(player.nick)

            elif command == "watch":
                if len(args) > 0:
                    out =  self.watch(self.get_target_player(player, args[0]))
                else:
                    out =  self.watch(player)

            elif command == "balance":
                if len(args) > 0:
                    out =  self.balance(self.get_target_player(player, args[0]))
                else:
                    out =  self.balance(player)

            elif command == "mine":
                out = self.mine(player)

            elif command == "transfer":
                if len(args) < 2:
                    out =  "transfer <target> <amount>"
                else:
                    out =  self.transfer(player, self.get_target_player(player, args[0]), int(args[1]))

            elif player.health <= 0:
                out = "You cannot do that while unconscious!"

# ----- commands that require consciousness below -----

            elif command == "steal":
                if len(args) < 2:
                    out =  "steal <target> <amount>  - rolls a d6. If <3, you steal <target> <amount> coins. Otherwise, you pay a <amount> * 2 fine to "+ self.bot_nick + "."
                else:
                    out =  self.steal(player, self.get_target_player(player, args[0]), int(args[1]))

            elif command == "frame":
                if len(args) < 2:
                    out =  "frame <target> <amount>  - rolls a d6. If <3, you make <target> pay a fine of <amount> coins to " + self.bot_nick + ". Otherwise, you pay a ceil(<amount>/2) to <target and floor(<amount>/2) to " + self.bot_nick + " as fines."
                else:
                    out =  self.frame(player, self.get_target_player(player, args[0]), int(args[1]))

            elif command == "punch":
                if len(args) < 1:
                    out =  "punch <target>"
                else:
                    out =  self.punch(player, self.get_target_player(player, args[0]))

            elif command == "roll":
                if len(args) < 1:
                    out =  "roll <amount> - rolls a d100 against watcoinbot. result<50 wins <amount>, result >=50 loses <amount>"
                else:
                    out =  self.roll(player, int(args[0]))

            elif command == "invent":
                if len(args) < 2:
                    out = "invent <itemname> <price> - invent an item called <itemname> which can be bought for <price>. An invention costs 100 coins."
                else:
                    out = self.invent(player, args[0], int(args[1]))

            elif command == "create":
                if len(args) < 2:
                    out = "create <itemname> <count> - create <count> <itemname>s. You have to pay the price and must be the inventor of the item!"
                else:
                    out = self.create(player, args[0], int(args[1]))

            elif command == "give":
                if len(args) < 3:
                    out = "give <target> <count> <itemname>"
                else:
                    out = self.give(player, args[0], int(args[1]), args[2])

            else:
                out = None
                                
            return out
            
        except:
            return "wat?"
            
    def get_target_player(self, player, target_nick):
        if target_nick == player.nick:
            return player
        elif target_nick == self.bot_nick:
            return self.bot_player
        else:
            return WatbotPlayer(self.db, target_nick)

    def watch(self, player):
        out = ( 
            "Watting: " + str(player.watting) + "(" + str(player.watting_exp) + ") / " + 
            "Anarchy: " + str(player.anarchy) + "(" + str(player.anarchy_exp) + ") / " + 
            "Trickery: " + str(player.trickery) + "(" + str(player.trickery_exp) + ") " +
            "Coins: " + str(player.coins) + " " + 
            "Health: " + str(player.health) 
        )
        return out 

    def balance(self, player):
        out = player.nick + "'s watcoin balance is " + str(player.coins) + ". Mining time credit: " + self.dhms(int(self.now - player.last_mine)) + " seconds."
        return out

    def mine(self, player):
        delta = self.now - player.last_mine
      
        if delta < 3600: 
            return "wat? not so soon again!"
         
        if delta < 36000:
            mined_coins = int(delta / 3600)
        elif delta < 86400:
            mined_coins = 10
        elif delta < 2592000:
            mined_coins = 50
        else:
            mined_coins = 1000

        player.coins += mined_coins
        player.last_mine = self.now

        out = player.nick + " mined " + str(mined_coins) + " coins for " + self.dhms(int(delta)) + " seconds and now has " + str(player.coins) + " watcoins."

        player.update(out)
        return out


    def transfer(self, player, target_player, amount):
        if amount < 0:
            return "wat? you thief!"

        if player.coins < amount:
            return "wat? you poor fuck don't have enough!"

        player.coins -= amount
        target_player.coins += amount

        if amount != 1:   
            out = player.nick + " sent " + target_player.nick + " " +str(amount) + " watcoins."
        else:
            out = player.nick + " sent " + target_player.nick + " a watcoin."

        player.update(out)
        target_player.update(out)

        return out

    def mega_wat(self, player):

        mega_number = random.randint(1,1000000)
        kilo_number = random.randint(1,1000)

        print "mega_wat(" + player.nick + ") mega_number == " + str(mega_number) + ", kilo_number == " + str(kilo_number)
      
        out = None
      
        if mega_number == 23:
            player.coins += 1000000
            out = "OMGWATWATWAT!!!! " + player.nick + " has won the MegaWat lottery and gains 1000000 watcoins!"
         
        if kilo_number == 5:
            player.coins += 1000
            out = "OMGWAT! " + player.nick + " has won the KiloWat lottery and gains 1000 watcoins!"

        player.watting_exp += 1        
        player.update(out)
        
        return out
        
    def roll(self, player, amount):
        if amount < 0:
            return "wat? nonono!"
            
        if player.coins < amount:
            return "wat? you broke, go away!"

        if self.bot_player.coins < amount:
            bot_mining_delta = self.now - self.bot_player.last_mine
            if bot_mining_delta > 86400:
                return self.bot_nick + " doesn't have enough coins for this, but " + self.bot_nick + " can mine! " + self.mine(self.bot_player) + self.roll(player, amount)
            else: 
                return "wat? " + self.bot_nick + " only has " + str(bot_player.coins) + " wtc left. Try again later or beg someone to fund the bot. " + self.bot_nick + " will mine in " + str(self.dhms(int(86400 - bot_mining_delta))) + "."
      
        number = random.randint(1, 100)
      
        if number < 50:
            player.coins += amount
            player.trickery_exp += 1
            self.bot_player.coins -= amount
            out = player.nick + " rolls a d100 (<50 wins): " + str(number) + ". You win! Your new balance is " + str(player.coins) + "."
        else:
            player.coins -= amount
            self.bot_player.coins += amount
            out = player.nick + " rolls a d100 (<50 wins): " + str(number) + ". You lose. Your new balance is " + str(player.coins) + "."

        player.update(out)
        self.bot_player.update(out)
        return out

    def invent(self, player, itemname, price):
        if price <= 0:
            return "wat? nonono!"

        invent_cost = 100 - player.watting
            
        if player.coins < invent_cost:
            return "wat? inventions cost you " + str(invent_cost) + " coins, but you're poor!"

        if self.db.invent_item(itemname, player.nick, price):
            player.coins -= invent_cost
            self.bot_player.coins += invent_cost
            out = player.nick + " invented " + itemname + " (" + str(price) + ")."
            player.update(out)
            self.bot_player.update(out)
        else:
            out = "wat?" + itemname + " already invented!"

        return out

    def create(self, player, itemname, count):
        if count <= 0:
            return "wat? nonono!"

        (itemname, inventor_nick, price) = self.db.get_item(itemname)
        
        if player.nick.lower() != inventor_nick:
            return "wat? that's not your invention!"

        if count * price > player.coins:
            return "wat? you need more money for that!"
            
        player.coins -= count * price
        original_count = self.db.get_item_count(player.nick, itemname)
        self.db.set_item_count(player.nick, itemname, original_count + count)
        
        out = player.nick + " created " + str(count) + " " + itemname + "."
        player.update(out)
        return out

    def give(self, player, target_nick, count, itemname):
        player_item_count = self.db.get_item_count(player.nick, itemname)
        if player_item_count < count:
            return "wat? you don't have that many!"
            
        self.db.set_item_count(player.nick, itemname, player_item_count - count)
        
        target_item_count = self.db.get_item_count(target_nick, itemname)
        self.db.set_item_count(target_nick, itemname, target_item_count + count)
        
        return player.nick + " gave " + target_nick + " " +  str(count) + " " + itemname

    def steal(self, player, target_player, amount):
        if amount < 0:
            return "wat? nonono!"
            
        if player.coins < amount * 2:
            return "wat? you don't have enough to pay the possible fine!"
      
        if target_player.coins < amount:
            return "wat? " + target_player.nick + " is a poor fuck and doesn't have " + str(amount) + " coins."
         
        number = random.randint(1,6)

        if number <3:
            out = player.nick + " rolls a d6 to steal " + str(amount) + " watcoins from " + target_player.nick + ": " + str(number) + " (<3 wins). You win! Sneaky bastard!"
            player.coins += amount
            player.anarchy_exp += 1
            target_player.coins -= amount
            player.update(out)
            target_player.update(out)
        else:
            out = player.nick + " rolls a d6 to steal " + str(amount) + " watcoins from " + target_player.nick + ": " + str(number) + " (<3 wins). You were caught and pay " + str(2 * amount) + " coins to " + self.bot_nick + "."
            player.coins -= 2 * amount
            self.bot_player.coins += 2 * amount
            player.update(out)
            self.bot_player.update(out) 
      
        return out
         
    def frame(self, player, target_player, amount):
        if amount < 0:
            return "wat? nonono!"
            
        if player.coins < amount:
            return "wat? you don't have enough to pay the possible fine!"
      
        if target_player.coins < amount:
            return "wat? " + target_player.nick + " is a poor fuck and doesn't have " + str(amount) + " coins."
         
        number = random.randint(1,6)

        if number <3:
            out = player.nick + " rolls a d6 to frame " + str(amount) + " watcoins from " + target_player.nick + ": " + str(number) + " (<3 wins). You win! " + target_player.nick + " pays " + str(amount) + " to " + self.bot_nick + "."

            player.anarchy_exp += 1             
            target_player.coins -= amount
            self.bot_player.coins += amount
            player.update(out)
            target_player.update(out)
            self.bot_player.update(out)
        else:
            target_amount = int(math.ceil(float(amount)/2))
            bot_amount = int(math.floor(float(amount)/2))
            out = player.nick + " rolls a d6 to frame " + str(amount) + " watcoins from " + target_player.nick + ": " + str(number) + " (<3 wins). You were caught and pay " + str(target_amount) + " coins to " + target_player.nick + " and " + str(bot_amount) + " coins to " + self.bot_nick + "."
            
            player.coins -= amount
            target_player.coins += target_amount
            self.bot_player.coins += bot_amount
            
            player.update(out)
            target_player.update(out)
            self.bot_player.update(out)
      
        return out
      
    def punch(self, player, target_player):

        number = random.randint(1, 6)
        dmg = random.randint(1, 6)

        if number <3:
            dmg += player.anarchy
            out = player.nick + " rolls a d6 to punch " + target_player.nick + ": " + str(number) +". " + player.nick + " hits for " + str(dmg) + " points of damage."
            target_player.health -= dmg
            if target_player.health <= 0:
                out += " " + target_player.nick + " falls unconscious."
            target_player.update(out)
        else:
            dmg += target_player.anarchy
            out = player.nick + " rolls a d6 to punch " + target_player.nick + ": " + str(number) +". " + player.nick + " misses. " + target_player.nick + " punches back for " + str(dmg) + " points of damage."
            player.health -= dmg
            if player.health <= 0:
                out += " " + player.nick + " falls unconscious."
            player.update(out)
      
        return out

    def dhms(self, seconds):
        days = int(seconds / 86400)
        seconds -= days * 86400
        hours = int(seconds / 3600)
        seconds -= hours * 3600
        minutes = int(seconds / 60)
        seconds -= minutes * 60
      
        out = ""
        show = False
      
        if days > 0:
            out += str(days) + "d "
            show = True
         
        if hours > 0 or show:
            out += str(hours) + "h "
            show = True
         
        if minutes > 0 or show:
            out += str(minutes) + "m "
            show = True
         
        out += str(seconds) + "s"
        return out

