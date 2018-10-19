#!/usr/bin/python

# LICENSE: DO WAT YOU WANT WITH THIS!


from watbot_config import WatbotConfig
from watbot_db import WatbotDB
from watbot_game import WatbotGame
from watbot_irc import WatbotIRC

db   = WatbotDB(WatbotConfig)
game = WatbotGame(WatbotConfig, db)
irc  = WatbotIRC(WatbotConfig, game)

irc.main_loop()
