# State class and helpers for music.py
from configstartup import config
from commands.playing import CURRENT_PRESENCE
from helpers import *

import os
import isodate
import queue
import random
import multiprocessing as mp

from discord import Game, Colour
from requests import get, Session
from requests.exceptions import RequestException
from bs4 import BeautifulSoup
# YouTube API things. Required for all YouTube links, searches, etc.
# Source: https://github.com/youtube/api-samples/blob/master/python/search.py
# Set environment variable to the API key value
# from the APIs & auth > Registered apps tab of https://cloud.google.com/console
# Please ensure that you have enabled the YouTube Data API for your project.
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
DEVELOPER_KEY = config['KEYS'].get('YouTube')
YOUTUBE_API_SERVICE_NAME = 'youtube'
YOUTUBE_API_VERSION = 'v3'

API_LIMIT = 50
GEO_REGION = "AU"
PAUSE_UTF = "\u23F8 "
PLAY_UTF = "\u25B6 "
REPEAT_LIST_UTF = "\U0001F501"
REPEAT_SONG_UTF = "\U0001F502"
SAMPLE_RATE = 48000
SUGG_CLASS = " content-link spf-link yt-uix-sessionlink spf-link "
VID_PREFIX = "https://www.youtube.com/watch?v="
YT_PREFIX = "https://www.youtube.com"

# Singleton pattern for state


class State:
    class __State:
        __auto = True
        __channel = None
        __list_limit = 10
        __lock = False
        __paused = False
        __player = None
        __playlist = []
        __presence = ""
        __q = mp.Queue()
        __repeat = "none"
        __session = None
        __voice = None
        __volume = float(7)
        running = False

        def __init__(self):
            pass

        def printState(self):
            out = "-----------------\nState:"
            out += "\nauto: " + str(self.__auto)
            out += "\nchannel: " + str(self.__channel)
            out += "\nlist_limit: " + str(self.__list_limit)
            out += "\npaused: " + str(self.__paused)
            out += "\nplayer: " + str(self.__player)
            out += "\nplaylist: " + str(self.__playlist)
            out += "\npresence: " + str(self.__presence)
            out += "\nrepeat: " + str(self.__repeat)
            out += "\nvolume: " + str(self.__volume)
            out += "\nvoice: " + str(self.__voice)
            out += "\nrunning: " + str(self.running)
            print(out)

        def addSong(self, song):
            self.__playlist.append(song)
            return bold("Added:") + " [%s] %s" % \
                (duration(song['duration']), song['title'])

        def addList(self, lis):
            self.__playlist.extend(lis)
            out = bold("Added Songs:") + "\n"
            for elm in lis:
                d = duration(elm['duration'])
                out += bold("[%s] %s" % (d, elm['title']))
            return out

        def beginSession(self): self.__session = Session()

        def cleanSession(self):
            self.__session.close()
            self.__session = None

        def checkListIndex(self, index):
            try:
                index = int(index)
            except ValueError:
                raise CommandFailure("Please use a number!")
            if len(self.__playlist) == 0:
                raise CommandFailure("Playlist is Empty!")
            if index >= len(self.__playlist):
                raise CommandFailure("Index out of playlist range")
            return index

        def freeLock(self): self.__lock = False

        def getChannel(self): return self.__channel

        def getListLimit(self): return self.__list_limit

        def getSession(self): return self.__session

        def getSong(self, index=0):
            index = self.checkListIndex(index)
            return self.__playlist[index]

        def getNext(self):
            if len(self.__playlist) > 0:
                return self.__playlist[0]
            else:
                return None

        def getPlaylist(self): return self.__playlist

        def getPresence(self): return self.__presence

        def getVoice(self): return self.__voice

        def handlePop(self, client):
            # Pre: playlist is not empty
            out = None
            if self.__repeat == "song":
                pass
            elif self.__repeat == "list":
                self.__playlist.append(self.__playlist.pop(0))
            else:
                song = self.__playlist.pop(0)
                if self.__auto and self.isListEmpty():
                    url = song['webpage_url']
                    mp_call(auto_info, url, song['author'])
                    self.lock()

        def hasPlayer(self): return True if self.__player else False

        def isAuto(self): return self.__auto

        def isDone(self):
            try:
                return self.__player.is_done()
            except AttributeError:
                return True

        def isListEmpty(self): return len(self.__playlist) == 0

        def isLocked(self): return self.__lock

        def isPlaying(self):
            try:
                return self.__player.is_playing() or self.__paused
            except AttributeError:
                return False

        def lock(self): self.__lock = True

        def toggleAuto(self):
            self.__auto = not self.__auto
            return "on" if self.__auto else "off"

        def playerDuration(self): return duration(self.__player.duration)

        def playerTitle(self): return self.__player.title

        def qGet(self):
            try:
                return self.__q.get_nowait()
            except queue.Empty:
                return None

        def qPut(self, song):
            try:
                self.__q.put_nowait(song)
            except queue.Full:
                CommandFailure("Multiprocessing queue full!")

        def resetSession(self):
            self.__session.close()
            self.__session = Session()
            return "HTTP Session reset!"

        def setAuto(self, auto): self.__auto = auto

        def setChannel(self, channel): self.__channel = channel

        def setLimit(self, limit):
            self.__list_limit = limit
            if 0 < limit <= API_LIMIT:
                list_limit = limit
            else:
                raise CommandFailure("Please enter an integer in [0-50]")
            return "Playlist fetch limit set to: %d" % limit

        def stat(self):
            col = Colour.red() if self.__paused else Colour.green()
            sta = " Now "
            sta += "Paused" if self.__paused else "Playing"
            sta += ": [%s] %s" % (self.playerDuration(), self.playerTitle())
            ti = ""
            if len(self.__playlist) > 1:
                ti = "Up Next:"
            au = "on" if self.__auto else "off"
            fo = "Repeat: %s | Auto: %s" % (self.__repeat, au)
            return sta, col, ti, fo

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
            index = self.checkListIndex(index)
            song = self.__playlist.pop(index)

            # Kill the player if we remove the currently playing song
            if index == 0:
                self.stop()
            return bold("Removed: [%s] %s" %
                        (duration(song['duration']), song['title']))

        def repeat(self, mode):
            if mode.lower() == "song":
                if self.__repeat != "none":
                    self.__presence = self.__presence[1:]
                self.__presence = REPEAT_SONG_UTF + self.__presence
                self.__repeat = mode.lower()

            elif mode.lower() == "list":
                if self.__repeat != "none":
                    self.__presence = self.__presence[1:]
                self.__presence = REPEAT_LIST_UTF + self.__presence
                self.__repeat = mode.lower()

            elif mode.lower() == "none":
                if self.__repeat != "none":
                    self.__presence = self.__presence[1:]
                self.__repeat = mode.lower()

            else:
                raise CommandFailure("Use `none`, `song`, or `list`!")

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

        def shuffle(self):
            if len(self.__playlist) == 0:
                raise CommandFailure("Playlist empty!")
            random.shuffle(self.__playlist)
            return "Shuffled Playlist!"

        def stop(self):
            try:
                self.__player.stop()
            except AttributeError:
                pass

        def volume(self, *lvl):
            if not self.hasPlayer():
                raise CommandFailure("Nothing playing!")

            if len(lvl) == 0:
                return self.__volume

            lvl = " ".join(lvl)

            try:
                lvl = float(lvl)
                if 0 <= lvl <= 100:
                    self.__volume = lvl
                    self.__player.volume = lvl/100
                    return "Volume changed to %f%%" % lvl
            except ValueError:
                raise CommandFailure("Please use a number between 0-100")

        async def clean(self, client):
            if self.__player:
                self.__player.stop()
                self.__player = None
            self.__playlist.clear()
            self.__presence = CURRENT_PRESENCE
            await self.updatePresence(client)

        async def embed(self, client, emb):
            await client.send_message(self.__channel, embed=emb)

        async def message(self, client, msg):
            await client.send_message(self.__channel, msg)

        async def joinVoice(self, client, message):
            channel = message.author.voice.voice_channel
            if not channel:
                raise CommandFailure("Please join a voice channel first")

            vclients = list(client.voice_clients)
            voices = [x.server for x in vclients]
            out = None
            try:
                voice = vclients[voices.index(message.server)]
            except ValueError:  # Not connected, join a vc
                voice = await client.join_voice_channel(channel)
                voice.encoder_options(sample_rate=SAMPLE_RATE, channels=2)
                self.setChannel(message.channel)

                out = "Joined %s, " % code(channel.name)
                out += "Binding to %s" % chan(self.getChannel().id)
            self.__voice = voice
            return voice, out

        async def playNext(self):
            song = self.getNext()
            if song == None:
                return None
            url = song['webpage_url']
            opts = {'format': 'bestaudio[ext=m4a]'}
            # https://github.com/Rapptz/discord.py/issues/315
            # Make sure you have ffmpeg-3 or above
            beforeArgs = "-reconnect 1 -reconnect_streamed 1 \
                    -reconnect_delay_max 5"
            self.__player = await self.__voice.create_ytdl_player(
                url,
                ytdl_options=opts,
                before_options=beforeArgs
            )
            self.__player.start()
            self.__player.volume = self.__volume/100

            presence = PLAY_UTF + song['title']
            if self.__repeat == "song":
                presence = REPEAT_SONG_UTF + presence
            if self.__repeat == "list":
                presence = REPEAT_LIST_UTF + presence
            self.__presence = presence

            return bold("Now Playing:") + " [%s] %s" % \
                (self.playerDuration(), self.playerTitle())

        async def updatePresence(self, client):
            await client.change_presence(game=Game(name=self.__presence))

    instance = __State()

    def __init__(self):
        pass


def mp_call(func, *args):
    # A non-blocking process is spawned
    # Not able to pass up return values, use mp.Queue()
    print("Spawning new process for " + str(func))
    p = mp.Process(target=func, args=args)
    p.start()


def auto_info(url, author):  # Expensive
    content = None
    try:
        resp = State.instance.getSession().get(url, stream=True)
        if is_good_response(resp):
            content = resp.content
        else:
            raise CommandFailure("Bad Response from %s" % url)
    except RequestException as e:
        raise CommandFailure(
            "Error during requests to %s : %s" % (url, str(e)))
    try:
        html = BeautifulSoup(content, 'lxml')
    except BadHTMLError as e:
        raise CommandFailure(e.message)
    a = html.find('a', class_=SUGG_CLASS)
    if a is not None:
        info = video_info(YT_PREFIX + a['href'], author)
        State.instance.qPut(info)
    else:
        # Could not get suggestion, try again
        State.instance.freeLock()


def check_bot_join(client, message):
    voices = [x.server for x in list(client.voice_clients)]
    try:
        v_index = voices.index(message.server)
    except ValueError:
        raise CommandFailure("Bot is not in a VC yet!")


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
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s'
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
    info['thumb'] = vid['snippet']['thumbnails']['default']['url']
    info['author'] = author
    return info


def playlist_info(url, author):  # Expensive
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
            maxResults=State.instance.getListLimit()
        ).execute()
    except HttpError as e:
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s'
              % (timestamp(), e.resp.status, e.content))
        return None

    info = dict()
    for video in videos['items']:
        # MP Create a new dict each iteration and append to mp queue
        copy = info.copy()
        copy = video_info(video['contentDetails']['videoId'], author)
        State.instance.qPut(copy)


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
        print('%s YOUTUBE: A HTTP error %d occurred:\n%s'
              % (timestamp(), e.resp.status, e.content))
        return None

    try:
        return video_info(search_response['items'][0]['id']['videoId'], author)
    except IndexError:
        raise CommandFailure(bold("Couldn't find %s" % query))
