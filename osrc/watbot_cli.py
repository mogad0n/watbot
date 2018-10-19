#!/usr/bin/python

from watbot_config import WatbotConfig
from watbot_db import WatbotDB
from watbot_game import WatbotGame
from watbot_console import WatbotConsole

import sys

if len(sys.argv) > 1:

  db   = WatbotDB(WatbotConfig)
  game = WatbotGame(WatbotConfig, db)
  con  = WatbotConsole(WatbotConfig, game, sys.argv[1])

  con.main_loop()
else:
  print "usage: " + sys.argv[0] + " <nickname>"
  