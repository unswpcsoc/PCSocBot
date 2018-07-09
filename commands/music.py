from commands.base import Command
from helpers import *

from collections import deque

from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

import discord, asyncio, datetime, youtube_dl, os

# Source: https://github.com/youtube/api-samples/blob/master/python/search.py
# Set DEVELOPER_KEY to the API key value from the APIs & auth > Registered apps
# tab of
#   https://cloud.google.com/console
# Please ensure that you have enabled the YouTube Data API for your project.
# TODO make api key environment variable
DEVELOPER_KEY = os.environ['YT_API']
YOUTUBE_API_SERVICE_NAME = 'youtube'
YOUTUBE_API_VERSION = 'v3'
MAX_RESULTS = 5

SLEEP_INTERVAL = 1
GEO_REGION = "AU"
SAMPLE_RATE = 48000

PAUSE_UTF = "\u23F8"
PLAY_UTF = "\u25B6"

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

    async def eval(self, *args):
        global bind_channel
        global player
        global playlist

        args = " ".join(args)

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

        if args.startswith("http"):
            # URL
            url = args
        else:
            # Not a URL, search youtube
            try:
                vid = youtube_search(args)
                url = "https://www.youtube.com/watch?v=%s" % vid
            except HttpError as e:
                print('An HTTP error %d occurred:\n%s' \
                        % (e.resp.status, e.content))
                return "Something went wrong!"

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

    async def eval(self):
        global bind_channel
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Paused Playing: [%s] %s" % (duration, player.title))

        # Change presence
        presence = PAUSE_UTF + player.title
        await self.client.change_presence(game=discord.Game(name=presence))

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

        # Change presence
        presence = PLAY_UTF + player.title
        await self.client.change_presence(game=discord.Game(name=presence))

        return out


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

    # Set bitrate
    voice.encoder_options(sample_rate=SAMPLE_RATE, channels=2)

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
                # Play the next song in the list
                song = playlist.pop(0)
                url = song['webpage_url']

            except IndexError:
                # Nothing in playlist, break
                out = bold("Stopped Playing")
                await client.send_message(bind_channel, out)
                # Reset presence
                await client.change_presence(game=None)
                # Reset player
                player = None
                # Exit event loop
                break

            player = await voice.create_ytdl_player(url)
            player.start()
            player.volume = volume/100  # "That's how you get tinnitus"

            # Print the message in the supplied channel
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            presence = song['title']
            out = bold("Now playing: [%s] %s" % (duration, presence))
            await client.send_message(bind_channel, out)

            # Change presence to the currently playing song
            presence = PLAY_UTF + presence
            await client.change_presence(game=discord.Game(name=presence))

        await asyncio.sleep(SLEEP_INTERVAL)


def youtube_search(query):
    youtube = build(YOUTUBE_API_SERVICE_NAME, YOUTUBE_API_VERSION, 
            developerKey=DEVELOPER_KEY)

    # Call the search.list method to retrieve results matching the specified
    # query term.
    search_response = youtube.search().list(
        q=query,
        part='id,snippet',
        maxResults=MAX_RESULTS,
        regionCode=GEO_REGION,
        type='video'
        ).execute()

    videos = []

    # Add each result to the appropriate list, and then display the lists of
    # matching videos, channels, and playlists.
    for search_result in search_response.get('items', []):

        #if search_result['id']['kind'] == 'youtube#video':
        videos.append(search_result['id']['videoId'])
            
    return videos[0]

