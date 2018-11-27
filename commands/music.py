from commands.base import Command
from commands.playing import CURRENT_PRESENCE
from helpers import *

from discord import Game, Embed, Colour
import asyncio, datetime, isodate, youtube_dl, os, random
from requests import get
from requests.exceptions import RequestException
from contextlib import closing
from bs4 import BeautifulSoup

# YouTube API things. Required for all YouTube links, searches, etc.
# Source: https://github.com/youtube/api-samples/blob/master/python/search.py
# Set environment variable to the API key value from the APIs & auth > Registered apps
# tab of
#   https://cloud.google.com/console
# Please ensure that you have enabled the YouTube Data API for your project.
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
DEVELOPER_KEY = os.environ['YT_API']
YOUTUBE_API_SERVICE_NAME = 'youtube'
YOUTUBE_API_VERSION = 'v3'

# Global constants
SLEEP_INTERVAL = 1
GEO_REGION = "AU"
SAMPLE_RATE = 48000
VID_PREFIX = "https://www.youtube.com/watch?v="
PLIST_PREFIX = "https://www.youtube.com/playlist?list="
WRONG_PLIST = "https://www.youtube.com/watch?list="
YDL_SITES = "https://rg3.github.io/youtube-dl/supportedsites.html"
PAUSE_UTF = "\u23F8 "
PLAY_UTF = "\u25B6 "
REPEAT_LIST_UTF = "\U0001F501"
REPEAT_SONG_UTF = "\U0001F502"
DC_TIMEOUT = 300
SUGG_CLASS = " content-link spf-link yt-uix-sessionlink spf-link "
YT_PREFIX = 'https://www.youtube.com'

# Global vars
#auto = True
#bind_channel = None
#list_limit = 10
#paused = False
#player = None
playlist = []
#presence = CURRENT_PRESENCE
repeat = "none"
#volume = float(15)

# Singleton pattern for state
class State:
    class __State:
        __auto = True
        __channel = None
        __list_limit = 10
        __paused = False
        __player = None
        __playlist = []
        __presence = CURRENT_PRESENCE
        __repeat = "none"
        __volume = float(15)
        __was_playing = False
        running = False

        def __init__(self):
            pass

        def addSong(self, song):
            self.__playlist.append(song)
            return bold("Added:") + " [%s] %s" % \
                    (duration(song['duration']), song['title'])

        def checkListIndex(self, index):
            try:
                index = int(index)
            except ValueError:
                raise CommandFailure("Please use a number!")

            if len(self.__playlist) == 0:
                raise CommandFailure("Playlist is Empty!")
            
            if index >= len(self.__playlist):
                raise CommandFailure("Index out of playlist range")

        def cleanPlayer(self):
            try:
                self.__player.stop()
                self.__player = None
                self.__was_playing = False
            except AttributeError:
                pass

        def getAuto(self):
            return self.__auto

        def getChannel(self):
            return self.__channel

        def getListLimit(self):
            return self.__list_limit

        def getSong(self, index):
            self.checkListIndex(index)
            return self.__playlist[index]

        def getNext(self):
            if len(self.__playlist) > 0:
                return self.__playlist[0]
            else:
                return None

        def getPresence(self):
            return self.__presence

        def handlePop(self):
            if repeat == "song": 
                pass
            elif repeat == "list": 
                playlist.append(playlist.pop(0))
            else: 
                playlist.pop(0) 

        def isPlaying(self):
            try:
                return self.__player.is_playing()
            except AttributeError:
                return False

        def isDone(self):
            try:
                return self.__player.is_done()
            except AttributeError:
                return True

        def isListEmpty(self):
            return len(self.__playlist) == 0

        def toggleAuto(self):
            self.__auto = not self.__auto
            return "on" if self.__auto else "off"

        def playerDuration(self):
            return duration(self.__player.duration)

        def playerTitle(self):
            return self.__player.title

        def setChannel(self, channel):
            self.__channel = channel

        def stat(self):
            col = Colour.red() if self.__paused else Colour.green()
            sta = "Paused" if self.__paused else "Playing"
            sta = "Now " + sta + ": [%s] %s" % \
                    (self.playerDuration(), self.playerTitle())
            ti = ""
            if len(self.__playlist) > 1:
                ti = "Up Next: (Repeat: %s)" % repeat 
            return out, sta, ti

        def pause(self):
            if not self.__player:
                raise CommandFailure("Not playing anything!")
            if self.__paused:
                raise CommandFailure("Already paused!")

            self.__player.pause()
            self.__paused = True

            self.__presence = PAUSE_UTF + self.__player.title
            if self.__repeat == "song": 
                self.__presence = REPEAT_SONG_UTF + self.__presence
            if self.__repeat == "list": 
                self.__presence = REPEAT_LIST_UTF + self.__presence

            return bold("Paused Playing:") + " [%s] %s" % \
                        (self.playerDuration(), self.playerTitle())

        def remove(self, index):
            self.checkListIndex(index)
            song = self.__playlist.pop(index)

            # Kill the player if we remove the currently playing song
            if index == 0: self.stop()
            return bold("Removed: [%s] %s" % \
                    (song['duration'], song['title']))

        def repeat(self, mode):
            presence = ""
            if mode.lower() == "song": 
                if self.__repeat != "none": presence = presence[1:]
                presence = REPEAT_SONG_UTF + presence
                self.__repeat = mode.lower()

            elif mode.lower() == "list": 
                if self.__repeat != "none": presence = presence[1:]
                presence = REPEAT_LIST_UTF + presence
                self.__repeat = mode.lower()

            elif mode.lower() == "none": 
                if self.__repeat != "none": presence = presence[1:]
                self.__repeat = mode.lower()

            else: raise CommandFailure("Not a valid argument!")

            self.__presence = presence
            return self.__repeat

        def resume(self):
            if not self.__player:
                raise CommandFailure("Not playing anything!")
            if not self.__paused:
                raise CommandFailure("Not paused!")

            self.__player.resume()
            self.__paused = False

            self.__presence = PLAY_UTF + self.__player.title
            if self.__repeat == "song": 
                self.__presence = REPEAT_SONG_UTF + self.__presence
            if self.__repeat == "list": 
                self.__presence = REPEAT_LIST_UTF + self.__presence

            return bold("Resumed Playing:") + " [%s] %s" % \
                        (self.playerDuration(), self.playerTitle())

        def stop(self):
            try: 
                self.__player.stop()
                self.__was_playing = False
            except AttributeError: 
                pass

        def volume(self, lvl):
            if not player:
                raise CommandFailure("Not playing anything!")

            try:
                lvl = float(lvl)
                if 0 <= lvl <= 100:
                    self.__volume = lvl
                    player.volume = lvl/100
                    return "Volume changed to %f%%" % lvl
            except ValueError:
                raise CommandFailure("Please use a number between 0-100")

        async def embed(self, client, emb):
            await client.send_message(self.__channel, embed=emb)

        async def message(self, client, msg):
            await client.send_message(self.__channel, msg)

        async def playNext(self):
            song = State.instance.getNext()
            if song == None: return None

            self.__player = await voice.create_ytdl_player(song['webpage_url'])
            self.__player.start()
            self.__player.volume = volume/100
            self.__was_playing = True

            presence = PLAY_UTF + song['title']
            if self.__repeat == "song": presence = REPEAT_SONG_UTF + presence
            if self.__repeat == "list": presence = REPEAT_LIST_UTF + presence
            self.__presence = presence

            return bold("Now Playing:") + " [%s] %s" % \
                    (self.playerDuration(), self.playerTitle())

        async def updatePresence(self, client):
            await client.change_presence(game=Game(name=self.__presence))

    instance = _State()

    def __init__(self):
        pass


class M(Command):
    desc = "Music"
    channels_required = []

class Auto(M):
    desc = "Toggles autoplay"

    async def eval(self):
        check_bot_join(self.client, self.message)

        out = "Autoplay is now "
        out += bold(State.instance.toggleAuto())
        await State.instance.message(self.client, out)

class Add(Auto):
    desc = "Adds the autoplay suggestion for a playlist index. Defaults \
            to the 0th."

    def eval(self, index=0):
        url = auto_get(State.getSong(index)['webpage_url'])
        return State.instance.addSong(video_info(url, self.client.user))

class Get(Auto):
    desc = "Gets the autoplay suggestion for a playlist index. Defaults \
            to the 0th."

    def eval(self, index=0):
        song = State.instance.getSong(index)
        item = video_info(auto_get(song['webpage_url']), self.client.user)

        out = bold("Got autosuggestion for") + " %s:" % (song['title'])
        out += "\n[%s] %s" % (duration(item['duration']), item['title'])
        out += "\nLink: " + noembed(item['webpage_url'])
        return out

class List(M):
    desc = "Lists the playlist."

    async def eval(self):
        if not State.instance.isPlaying():
            raise CommandFailure("Not playing anything!")

        status, colour, title = State.instance.stat()
        embed = Embed(title=title, colour=colour)
        embed.set_author(name=status)
        embed.set_footer(text="!m play [link/search]")
                        
        i = 0
        for song in playlist[1:]:
            i += 1
            title = "%d. [%s] %s" % \
                    (i, State.instance.playerDuration(), song['title'])
            embed.add_field(name=title, 
                            value="Added by: %s" % nick(song['author']), 
                            inline=False)

        await State.instance.embed(self.client, embed)

class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."

    async def eval(self): return await List.eval(self)

class ListLimit(M):
    desc = "Sets playlist fetch limit. Mods only."
    roles_required = [ "mod", "exec"]

    def eval(self, limit):
        global list_limit

        try: limit = int(limit)
        except ValueError: raise CommandFailure("Please enter a valid integer")

        # API limit is 50 videos
        if 0 < limit <= 50: list_limit = limit
        else: raise CommandFailure("Please enter an integer in [0-50]")

        return "Playlist fetch limit set to: %d" % limit

class Pause(M):
    desc = "Pauses music"

    async def eval(self):
        out = State.instance.pause()
        await State.instance.updatePresence()
        return out

class Play(M):
    desc = "Plays music. Binds commands to the channel invoked.\n"
    desc += "Accepts youtube links, playlists (first %d entries), and more!\n"
    desc += noembed(YDL_SITES) + "\n"
    desc += "Note: Playlists fetching through the YT API is limited to 50 vids"

    async def eval(self, *args):
        args = " ".join(args)

        channel = get_user_voice(self.message)

        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            # Get the voice channel
            v_index = voices.index(self.message.server)
            voice = vclients[v_index]
            was_connected = True
        except ValueError:
            # Not connected, join a vc
            voice, out = do_join(self.client, self.message)
            await State.instance.message(self.client, out)
            was_connected = False
            M.channels_required.append(State.instance.getChannel())

        if args.startswith("http"):
            url = args

            if url.startswith(PLIST_PREFIX):
                songs = playlist_info(url, self.message.author)
                playlist.extend(songs)

                out = bold("Added Songs:") + "\n"
                for song in songs:
                    d = duration(song['duration'])
                    out += bold("[%s] %s" % (str(d), song['title']))
                    out += "\n"

                await self.client.send_message(bind_channel, out)

            elif url.startswith(WRONG_PLIST):
                out = "Warning: invalid Playlist URL\n"
                out += "Please change `watch` to `playlist`"
                await State.instance.message(self.client, out)

            elif url.startswith(VID_PREFIX):
                # Video, could have playlist, add anyway
                song = video_info(url, self.message.author)
                out = State.instance.addSong(song)

                if len(url.split("list=")) > 1:
                    out += "\nYou have added a video that is "
                    out += "part of a playlist. Use this URL to add the "
                    out += "playlist:\n"
                    nu = PLIST_PREFIX + url.split("list=")[1].split("&")[0]
                    out += noembed(nu)

                await State.instance.message(self.client, out)

            else:
                try:
                    # Not a youtube link, use youtube_dl
                    # Get from youtube_dl
                    ydl_opts = {'geo_bypass_country': GEO_REGION}
                    with youtube_dl.YoutubeDL(ydl_opts) as ydl:
                        info = ydl.extract_info(url, download=False)

                    song = {}
                    song['webpage_url'] = info['webpage_url']
                    song['duration'] = info['duration']
                    song['title'] = info['title']
                    song['author'] = self.message.author

                    out = State.instance.addSong(song)
                    await State.instance.message(self.client, out)
                except:
                    # Unsupported URL
                    out = "Unsupported URL, see %s" % noembed(YDL_SITES)
                    raise CommandFailure(out)

        else:
            # Not a URL, search YouTube
            song = youtube_search(args, self.message.author)
            out = State.instance.addSong(song)
            await State.instance.message(self.client, out)

        # Music not running, start it
        if not State.instance.running:
            State.instance.running = True
            await music(voice, self.client, self.message.channel)

class Remove(M):
    desc = "Removes a song from the playlist. Defaults to the current song"

    def eval(self, index=0):
        check_bot_join(self.client, self.message)
        return State.instance.remove(index)

class Repeat(M):
    desc = "Toggle repeat for the current song or the whole playlist. "
    desc += "Accepted arguments are: 'none', `song` and `list`.\n"

    async def eval(self, mode):
        check_bot_join(self.client, self.message)
        out = State.instance.repeat(mode)
        await State.instance.updatePresence()
        return bold("Repeat mode set to: %s" % out)

class Resume(M):
    desc = "Resumes music"

    async def eval(self):
        out = State.instance.resume()
        await State.instance.updatePresence()
        return out

class Rm(M):
    desc = "See " + bold(code("!m") + " " + code("remove")) + "."

    def eval(self, pos=0): return Remove.eval(self, pos)

class Rp(M):
    desc = "See " + bold(code("!m") + " " + code("repeat")) + "."

    def eval(self, mode): return Repeat.eval(self.mode)

class Shuffle(M):
    desc = "Shuffles the playlist"

    def eval(self):
        global playlist

        # Check if connected to a voice channel
        check_bot_join(self.client, self.message)

        if playlist and len(playlist) > 1:
            random.shuffle(playlist)
        else:
            raise CommandFailure("No playlist!")

        return "Shuffled Playlist!"

class Sh(M):
    desc = "See " + bold(code("!m") + " " + code("shuffle")) + "."

    def eval(self): return Shuffle.eval(self)

class Skip(M):
    desc = "Skips the current song. Does not skip if repeat is `song`"

    def eval(self):
        check_bot_join(self.client, self.message)

        if not State.instance.isPlaying():
            raise CommandFailure("Not playing anything!")

        State.instance.stop()

        out = bold("Skipped: [%s] %s" % \
                (State.instance.playerDuration(), State.instance.playerTitle()))

        return out

class Stop(M):
    desc = "Stops playing but persists in voice. Also stops autoplaying."

    def eval(self):
        global bind_channel, playlist

        if not State.instance.isPlaying():
            return "Not playing anything!"

        State.instance.cleanPlayer()
        playlist.clear()
        auto = False

class Volume(M):
    desc = "Volume adjustment. Mods only."
    roles_required = [ "mod", "exec" ]

    def eval(self, level):
        return State.instance.volume(level)

class V(M):
    desc = "See " + bold(code("!m") + " " + code("volume")) + "."
    roles_required = [ "mod", "exec" ]

    def eval(self, level): return Volume.eval(self, level)

# HELPER FUNCTIONS #

def check_bot_join(client, message):
    voices = [ x.server for x in list(client.voice_clients) ]
    try:
        v_index = voices.index(message.server)
    except ValueError:
        raise CommandFailure("Bot is not in a VC yet!") 

def do_join(client, message):
    channel = message.author.voice.voice_channel
    if channel:
        voice = await client.join_voice_channel(channel)
        voice.encoder_options(sample_rate=SAMPLE_RATE, channels=2)
        State.instance.setChannel(message.channel)
        out = "Joined %s, " % code(channel.name)
        out += "Binding to %s" % chan(State.instance.getChannel().id)
        return voice, out
    raise CommandFailure("Please join a voice channel first")

def get_user_voice(message):
    channel = message.author.voice.voice_channel
    if not channel:
        raise CommandFailure("Please join a voice channel first")
    return channel

def video_info(url, author):
    youtube = build(YOUTUBE_API_SERVICE_NAME, YOUTUBE_API_VERSION, 
          developerKey=DEVELOPER_KEY)

    if url.startswith("http"):
        vid = url.split("v=")[1]
        vid = vid.split("&")[0]
    else:
        vid = url

    try:
        videos = youtube.videos().list(
                part='snippet, contentDetails',
                id=vid
                ).execute()
    except HttpError as e:
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s' \
                % (timestamp(), e.resp.status, e.content))
        return None

    try:
        vid = videos['items'][0]
    except IndexError:
        raise CommandFailure("Couldn't get info for %s" % url)

    info = {}
    info['title'] = vid['snippet']['title']
    info['webpage_url'] = VID_PREFIX + vid['id']
    duration = isodate.parse_duration(vid['contentDetails']['duration'])
    info['duration'] = duration.seconds
    info['author'] = author
    return info

def playlist_info(url, author):
    youtube = build(YOUTUBE_API_SERVICE_NAME, YOUTUBE_API_VERSION, 
          developerKey=DEVELOPER_KEY)

    if url.startswith("http"):
        vid = url.split("list=")[1]
        vid = vid.split("&")[0]
    else:
        vid = url

    try:
        videos = youtube.playlistItems().list(
                part='snippet, contentDetails',
                playlistId=vid,
                maxResults=list_limit
                ).execute()
    except HttpError as e:
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s' \
                % (timestamp(), e.resp.status, e.content))
        return None

    info_list = []
    info = dict()
    for video in videos['items']:
        # Create a new dict each iteration and append to list
        copy = info.copy()
        copy = video_info(video['contentDetails']['videoId'], author)
        info_list.append(copy)

    return info_list

def youtube_search(query, author):
    youtube = build(YOUTUBE_API_SERVICE_NAME, YOUTUBE_API_VERSION, 
            developerKey=DEVELOPER_KEY)

    try:
        search_response = youtube.search().list(
            q=query,
            part='id',
            maxResults=1,
            regionCode=GEO_REGION,
            type='video'
            ).execute()
    except HttpError as e:
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s' \
                % (timestamp(), e.resp.status, e.content))
        return None

    try:
        return video_info(search_response['items'][0]['id']['videoId'], author)
    except IndexError:
        raise CommandFailure(bold("Couldn't find %s" % query))

def auto_get(url):
    content = None
    try:
        with closing(get(url, stream=True)) as resp:
            if is_good_response(resp):
                content = resp.content
            else:
                raise CommandFailure("Bad Response from %s" % url)
    except RequestException as e:
        raise CommandFailure("Error during requests to %s : %s" % (url, str(e)))

    try:
        html = BeautifulSoup(content, 'html.parser')
    except BadHTMLError as e:
        raise CommandFailure(e.message)

    a = html.find('a', class_=SUGG_CLASS)
    result = YT_PREFIX + a['href']
    return result

# Big Boi
async def music(voice, client, channel):
    """ Music event loop
    This function takes a discord voice session, dicsord client, and channel.

    Note: Songs aren't popped when played; they are read off Playlist's head.
          We only pop a song when we need to access the next one.
          This allows us to toggle single-song repeat for the current song.
    """

    # sentinel vars
    dc_ticker = 0
    paused_dc = False

    # Begin event loop
    while True:

        # Handle autoplay
        if State.instance.isListEmpty() and State.instance.getAuto():
            suggestion = auto_get(playlist[0]['webpage_url'])
            result = video_info(suggestion, client.user)
            out = State.instance.addSong(result)
            await State.instance.message(client, out)

        # Handle player done
        if State.instance.isDone():
            if len(playlist) > 0:
                State.instance.handlePop()

                out = await State.instance.playNext(song['webpage_url'])
                if out = None: continue

                await State.instance.message(client, out)
                await State.instance.updatePresence(client)

            else:
                # Clean up
                State.instance.cleanPlayer()

                # Signal
                out = bold("Stopped Playing")
                await client.send_message(bind_channel, out)
                await client.change_presence(game=Game(\
                                            name=CURRENT_PRESENCE))

        # Handle no audience
        if len(voice.channel.voice_members) <= 1:   

            if dc_ticker >= DC_TIMEOUT: 
                State.instance.cleanPlayer()
                await voice.disconnect()

                out = "Timeout of [%s] reached," % duration(DC_TIMEOUT)
                out += " Disconnecting from %s," % code(voice.channel.name)
                out += " Unbinding from %s" % \
                        chan(State.instance.getChannel().id)

                M.channels_required.clear()
                await client.send_message(bind_channel, bold(out))
                await client.change_presence(game=Game(
                                            name=CURRENT_PRESENCE))

                State.instance.running = False
                return

            # Start counting
            if State.instance.isPlaying():
                paused_dc = True

                await client.change_presence(game=Game(name= \
                    State.instance.pause()))

                name = voice.channel.name
                out = bold("Nobody listening in %s, Pausing" % code(name))
                await State.instance.message(client)
                await State.instance.updatePresence(client)

            dc_ticker += SLEEP_INTERVAL

        # Someone joined, reset
        if len(voice.channel.voice_members) > 1 and paused_dc:
            paused_dc = False
            dc_ticker = 0

            State.instance.resume()
            name = voice.channel.name
            out = bold("Somebody has joined %s! Resuming" % code(name))
            await State.instance.message(client, out)
            await State.instance.updatePresence(client)

        await asyncio.sleep(SLEEP_INTERVAL)
