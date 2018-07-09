from commands.base import Command
from helpers import *

from collections import deque

import discord, asyncio, datetime, youtube_dl

SLEEP_INTERVAL = 1
GEO_REGION = "AU"

bind_channel = None
player = None
playlist = []
volume = float(12)


class M(Command):
    desc = "Music"
    channels_required = []


#class Join(M):
#    desc = "Binds the bot to the voice channel"
#
#    async def eval(self):
#        voice = await do_join(self.client, self.message)


class Leave(M):
    desc = "Boots the bot from voice channels"

    async def eval(self):
        # Checks if joined to a vc in the server
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            v_index = voices.index(self.message.author.server)
        except ValueError:
            raise CommandFailure("Not in a voice channel!")

        # Get the voice channel
        voice = vclients[v_index]

        channel = self.message.author.voice.voice_channel
        if channel:
            await voice.disconnect()
        else:
            raise CommandFailure("Please join a voice channel first")

        # Flush channels required
        M.channels_required.clear()
        return "Leaving %s, Unbinding from %s" % \
               (code(voice.channel.name), chan(bind_channel.id))


class Play(M):
    desc = "Plays music. Binds commands to the channel invoked."

    async def eval(self, *url):
        global bind_channel
        global player
        global playlist

        url = " ".join(url)

        # Check if user is connected to a vc
        channel = self.message.author.voice.voice_channel
        if not channel:
            raise CommandFailure("Please join a voice channel first")

        # Check if bot is connected already in the server
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            # Get the voice channel
            v_index = voices.index(self.message.server)
            voice = vclients[v_index]
        except ValueError:
            # Not connected, join a vc
            voice = await do_join(self.client, self.message)
            # Set channel required
            M.channels_required.append(bind_channel)

        # TODO Implement playlist scraping
        # TODO Implement searching
        # Get metadata
        ydl_opts = {'geo_bypass_country': GEO_REGION}
        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)

        song = {}
        song['webpage_url'] = info['webpage_url']
        song['duration'] = info ['duration']
        song['title'] = info ['title']
        playlist.append(song)

        # Send message acknowledging add
        duration = str(datetime.timedelta(seconds=int(song['duration'])))
        out = bold("Added: [%s] %s" % (duration, song['title']))
        await self.client.send_message(bind_channel, out)

        # Nothing is playing, start the music event loop
        if not player or player.is_done():
            await music(voice, self.client, self.message.channel)


class Pause(M):
    desc = "Pauses music"

    def eval(self):
        global bind_channel
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Paused Playing: [%s] %s" % (duration, player.title))
        return out


class Resume(M):
    desc = "Resumes music"

    async def eval(self):
        global bind_channel
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.resume()

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Resumed Playing: [%s] %s" % (duration, player.title))
        await self.client.send_message(bind_channel, out)


class Skip(M):
    desc = "Skips a song. Defaults to the current song."

    def eval(self, pos=-1):
        global bind_channel
        global player
        global playlist

        try:
            pos = int(pos)
        except ValueError:
            raise CommandFailure("Not a valid position!")

        if not player:
            raise CommandFailure("Not playing anything!")

        # Check if connected to a voice channel
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            v_index = voices.index(self.message.server)
        except ValueError:
            # Bot is not connected to a voice channel in this server
            raise CommandFailure("Please `!join` a voice channel first") 

        if 0 <= pos < len(playlist):
            # Remove the item from the playlist
            song = playlist.pop(pos)

            # Construct out message
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            out = bold("Removed: [%s] %s" % (duration, song['title']))

        elif pos == -1:
            # Construct out message
            duration = str(datetime.timedelta(seconds=int(player.duration)))
            out = bold("Removed: [%s] %s" % (duration, player.title))

            # Destroy player
            player.stop()
            player = None

        else:
            raise CommandFailure("Not a valid position!")

        return out


class Stop(M):
    desc = "Stops playing but persists in voice"

    def eval(self):
        global bind_channel
        global player
        global playlist

        if not player:
            return "Not playing anything!"

        player.stop()
        player = None
        playlist.clear()
        return bold("Stopped Playing")


class Volume(M):
    desc = "Volume adjustment"

    def eval(self, vol):
        global bind_channel
        global player
        global volume

        if not player:
            raise CommandFailure("Not playing anything!")

        try:
            vol = float(vol)
        except ValueError:
            raise CommandFailure("Please enter a number between 0-100")

        if 0 <= vol <= 100:
            # Change the global volume
            volume = vol
            player.volume = volume/100
            out = "Volume changed to %f%%" % vol
            return out
        else:
            raise CommandFailure("Please enter a number from 0-100")


class V(M):
    desc = "See " + bold(code("!m") + " " + code("volume")) + "."
    def eval(self, vol):
        return Volume.eval(self, vol)


class List(M):
    desc = "Lists the playlist"

    # TODO Embeds
    def eval(self):
        global bind_channel
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
            out += code("%d. [%s] %s" % (i, duration, song['title'])) 
            out += "\n"
            i += 1

        return out


class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."
    def eval(self):
        return List.eval(self)


async def do_join(client, message):
    global bind_channel

    # Checks if joined to a vc in the server
    vclients = list(client.voice_clients)
    voices = [ x.server for x in vclients ]
    if message.author.server in voices:
        return "Already joined a voice channel!"

    channel = message.author.voice.voice_channel
    if channel:
        voice = await client.join_voice_channel(channel)
    else:
        raise CommandFailure("Please join a voice channel first")

    # Bind the bot to the text channel it came from
    bind_channel = message.channel

    out = "Joined %s, " % code(channel.name)
    out += "Binding to %s" % chan(bind_channel.id)
    await client.send_message(bind_channel, out)

    return voice


async def music(voice, client, channel):
    global bind_channel
    global player
    global playlist
    global volume

    # TODO Make a proper event loop
    while True:

        # Poll when there is no player or the player is finished
        if not player or player.is_done():

            try:
                # Play the next song in the deque
                song = playlist.pop(0)
                url = song['webpage_url']
            except IndexError:
                # Nothing in playlist, break
                out = bold("Stopped Playing")
                await client.send_message(bind_channel, out)
                break

            # TODO Optimise first play scrape
            # TODO Presence
            player = await voice.create_ytdl_player(url)
            player.start()

            # Print the message in the supplied channel
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            out = bold("Now playing: [%s] %s" % (duration, song['title']))
            await client.send_message(bind_channel, out)

            # "That's how you get tinnitus"
            player.volume = volume/100

        await asyncio.sleep(SLEEP_INTERVAL)
