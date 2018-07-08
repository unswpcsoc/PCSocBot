from commands.base import Command
from helpers import *

from collections import deque

import discord, asyncio, datetime, youtube_dl

SLEEP_INTERVAL = 5

bind_channel = None
player = None
playlist = deque()
volume = 20


class M(Command):
    desc = "Music"


class Join(M):
    desc = "Binds the bot to the voice channel"

    async def eval(self):
        # Bind the bot to the text channel it came from
        global bind_channel
        bind_channel = self.message.channel

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

        return "Joined %s, Binding to %s" % (channel.name, bind_channel)


class Leave(M):
    desc = "Boots the bot from voice channels"

    async def eval(self):
        # Assume we are only running in one server
        # If you'd like to host the bot on multiple servers,
        # figure out how to handle it per server
        for vc in [ x for x in self.client.voice_clients ]:
            await vc.disconnect()

        return "Unbinding from %s" % bind_channel


class Play(M):
    desc = "Plays music. Must have `!join`ed before playing."

    async def eval(self, url):
        global player

        # Check if connected
        channel = self.message.author.voice.voice_channel
        if not channel:
            raise CommandFailure("Please join a voice channel first")

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

        # Get metadata
        ydl_opts = {}
        info = {}
        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)

        r_info = {}
        r_info['webpage_url'] = info['webpage_url']
        r_info['duration'] = info ['duration']
        r_info['title'] = info ['title']
        playlist.append(r_info)

        out = bold("Added: %s" % noembed(url))
        await self.client.send_message(bind_channel, out)

        # Check if nothing is playing and start the music event loop
        if not player or player.is_done():
            await music(voice, self.client, self.message.channel)


class Pause(M):
    desc = "Pauses music"

    def eval(self):
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()
        duration = str(datetime.timedelta(seconds=int(player.duration)))
        return bold("Paused Playing: [%s] %s" % (duration, player.title))


class Resume(M):
    desc = "Resumes music"

    def eval(self):
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.resume()
        duration = str(datetime.timedelta(seconds=int(player.duration)))
        return bold("Resumed Playing: [%s] %s" % (duration, player.title))

class Skip(M):
    desc = "Skips a song"

    def eval(self):
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        # Check if connected to a voice channel
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            v_index = voices.index(self.message.server)
        except ValueError:
            # Not connected to a voice channel in this server
            raise CommandFailure("Please `!join` a voice channel first") 

        # Destroy the player instance
        player.stop()
        player = None

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
        global volume

        if not player:
            raise CommandFailure("Not playing anything!")

        try:
            vol = float(vol)
        except ValueError:
            raise CommandFailure("Please enter a number between 0-100")

        if vol < 100 and vol > 0:
            # Change the global volume
            volume = vol
            player.volume = volume/100
            return "Volume changed to %f%%" % vol
        else:
            raise CommandFailure("Please enter a number between 0-100")


class List(M):
    desc = "Lists the playlist"

    def eval(self):
        global playlist

        if not player or player.is_done():
            raise CommandFailure("Not playing anything!")

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Now Playing: [%s] %s" % (duration, player.title))

        if playlist:
            out += "\n\n"
            out += bold("Up Next:")
            out += "\n"

        i = 0
        for song in playlist:
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            out += "%d. [%s] %s\n" % (i, duration, song['title']) 
            i += 1

        return out


class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."
    def eval(self)
    return List.eval(self)


async def music(voice, client, channel):
    global player
    global playlist
    global volume

    print("Entered music")
    while True:

        # Poll when there is no player or the player is finished
        if not player or player.is_done():
            print("Entered is_done")

            url = ''
            try:
                # Play the next song in the deque
                song = playlist.popleft()
                url = song['webpage_url']
            except IndexError:
                break

            player = await voice.create_ytdl_player(url)
            player.start()

            # Print the message in the supplied channel
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            out = bold("Now playing: [%s] %s" % (duration, song['title']))
            await client.send_message(channel, out)

            player.volume = volume

        await asyncio.sleep(SLEEP_INTERVAL)
