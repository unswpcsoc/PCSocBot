from commands.base import Command
from helpers import *

import discord

player = None


class M(Command):
    desc = "Music"


class Join(M):
    desc = "Binds the bot to the voice channel"

    async def eval(self):
        # Assume we are only running in one server
        # If you'd like to host the bot on multiple servers,
        # figure out how to handle it per server
        if self.client.voice_clients:
            return "Already joined a voice channel!"

        channel = self.message.author.voice.voice_channel
        if channel:
            voice = await self.client.join_voice_channel(channel)
        else:
            raise CommandFailure("Please join a voice channel first")

        return "Joined %s" % channel.name


class Leave(M):
    desc = "Boots the bot from voice channels"

    async def eval(self):
        # Assume we are only running in one server
        # If you'd like to host the bot on multiple servers,
        # figure out how to handle it per server
        for vc in [ x for x in self.client.voice_clients ]:
            await vc.disconnect()


class Play(M):
    desc = "Plays music. Must have `!join`ed before playing."

    async def eval(self, url):
        global player

        # Check if connected
        channel = self.message.author.voice.voice_channel
        if not channel:
            raise CommandFailure("Please join a voice channel first")

        if player:
            # TODO
            return "Already playing something"

        # Check if connected to a voice channel
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            v_index = voices.index(self.message.server)
        except ValueError:
            # Not connected to a voice channel in this server
            raise CommandFailure("Please `!join` a voice channel first") 

        # Get the voice channel
        voice = vclients[v_index]

        player = await voice.create_ytdl_player(url)
        player.start()
        return bold("Now Playing: [%s] %s" % (player.duration, player.title))


class Pause(M):
    desc = "Pauses music"

    def eval(self):
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()
        return bold("Paused Playing: [%s] %s" % (player.duration, player.title))


class Stop(M):
    desc = "Stops playing but persists in voice"

    def eval(self):
        global player

        if not player:
            return "Not playing anything!"

        player.stop()
        player = None
        return bold("Stopped Playing")


class Volume(M):
    desc = "Volume adjustment"

    def eval(self, vol):
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        try:
            vol = float(vol)
        except ValueError:
            raise CommandFailure("Please enter a number between 0-100")

        if vol < 100 and vol > 0:
            player.volume = vol/100
            return "Volume changed to %f%%" % vol
        else:
            raise CommandFailure("Please enter a number between 0-100")
