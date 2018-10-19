#!/usr/bin/python

import irc.client
import ssl

class WatbotIRC:
    """irc frontend class"""
    
    def __init__(self, config, game):

        self.config = config
        self.game   = game

        self.client = irc.client.IRC()
        server = self.client.server()

        if not config["ssl"]:
            server.connect(
                config["server"],
                config["port"],
                config["bot_nick"],
                username = config["bot_nick"],
                ircname = (config["prefix"] + " help"), 
            )
        else:
            ssl_factory = irc.connection.Factory(wrapper=ssl.wrap_socket)
            server.connect(
                config["server"],
                config["port"],
                config["bot_nick"],
                username = config["bot_nick"],
                ircname = (config["prefix"] + " help"), 
                connect_factory = ssl_factory
           )


        self.client.add_global_handler("welcome",    self.on_connect)
        self.client.add_global_handler("privmsg",    self.on_msg)
        self.client.add_global_handler("pubmsg",     self.on_msg)
        self.client.add_global_handler("disconnect", self.on_disconnect)


    def main_loop(self):
      self.client.process_forever()

    def on_connect(self, connection, event):
       if self.config["nickserv"]:
          connection.privmsg("nickserv", "identify " + self.config["password"])
          
       connection.join(self.config["channel"])

    def on_msg(self, connection, event):
        for a in event.arguments:
           self.do_command(connection, event.source, event.target, a)

    def do_command(self, connection, source, target, commandline):

       cl_list = commandline.strip().split(" ", 2)

       if len(cl_list) > 0:
          if cl_list[0].lower() == "wat":
             cl_list = [ self.config["prefix"], "wat" ]

       if len(cl_list) < 2:
          return

       if cl_list[0].lower() != self.config["prefix"]:
          return

       command = cl_list[1].lower()
       if len(cl_list) > 2:
          args_list = cl_list[2].split(" ")
       else:
          args_list = []

       if self.config["nickserv"]:
          # The @ must be there, irc standard
          host = source.split("@")[1]
          host_split = host.split("/")
          if len(host_split) < 3:
             return
          
          # Only allow nickserved users
          if host_split[0] != "tripsit":
             return      

       #nick = host_split[2]
       nick = source.split("!", 1)[0]   

       output = self.game.do_command(nick, command, args_list)

       
       if target == self.config["bot_nick"]:
          target = nick

       if not output is None:   
          connection.privmsg(target, output)
          
    def on_disconnect(self, connection, event):
       raise SystemExit()


