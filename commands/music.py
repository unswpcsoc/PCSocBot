from commands.base import Command
from helpers import *

from collections import deque

from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

from discord import Game, Embed, Colour
import asyncio, datetime, isodate, youtube_dl, os, random

# Source: https://github.com/youtube/api-samples/blob/master/python/search.py
# Set DEVELOPER_KEY to the API key value from the APIs & auth > Registered apps
# tab of
#   https://cloud.google.com/console
# Please ensure that you have enabled the YouTube Data API for your project.

# Make sure you set your environment variable
DEVELOPER_KEY = os.environ['YT_API']
YOUTUBE_API_SERVICE_NAME = 'youtube'
YOUTUBE_API_VERSION = 'v3'

# Max results for playlist
MAX_RESULTS = 10

SLEEP_INTERVAL = 1
GEO_REGION = "AU"
SAMPLE_RATE = 48000
VID_PREFIX = "https://www.youtube.com/watch?v="
PLIST_PREFIX = "https://www.youtube.com/playlist?list="
YDL_SITES = "https://rg3.github.io/youtube-dl/supportedsites.html"
PAUSE_UTF = "\u23F8"
PLAY_UTF = "\u25B6"

bind_channel = None
paused = False
player = None
playlist = []
volume = float(12)


class M(Command):
    desc = "Music"
    channels_required = []


class Leave(M):
    desc = "Boots the bot from voice channels"

    async def eval(self):
        global player
        global playlist

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

        # Clean player and playlist
        player = None
        playlist.clear()

        # Flush channels required
        M.channels_required.clear()
        return "Leaving %s, Unbinding from %s" % \
               (code(voice.channel.name), chan(bind_channel.id))


class Play(M):
    desc = "Plays music. Binds commands to the channel invoked.\n"
    desc += "Accepts youtube links, playlists (first %d entries), and more!\n"
    desc += noembed(YDL_SITES)

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

        if args.startswith("http"):
            # URL
            url = args

            # Scrape using yt API
            try:
                if url.startswith(PLIST_PREFIX):
                    # Pure playlist
                    songs = playlist_info(url, self.message.author)
                    playlist.extend(songs)

                    # Construct add message
                    out = bold("Added Songs:") + "\n"
                    for song in songs:
                        d = datetime.timedelta(seconds=int(song['duration']))
                        out += "[%s] %s\n" % (str(d), song['title'])

                    await self.client.send_message(bind_channel, out)

                elif url.startswith(VID_PREFIX):
                    # Video, could have playlist, add anyway
                    song = video_info(url, self.message.author)
                    playlist.append(song)

                    # Construct add message
                    d = str(datetime.timedelta(seconds=int(song['duration'])))
                    out = bold("Added:") + " [%s] %s" % (d, song['title'])

                    # Check for list param
                    if len(url.split("list=")) > 1:
                        out += "\nWarning: You have added a video that is "
                        out += "part of a playlist.Use this URL to add the "
                        out += "playlist:\n"
                        nu = PLIST_PREFIX + url.split("list=")[1].split("&")[0]
                        out += noembed(nu)

                    await self.client.send_message(bind_channel, out)

                else:
                    try:
                        # Not a youtube link, use youtube_dl
                        # Get from youtube_dl
                        ydl_opts = {'geo_bypass_country': GEO_REGION}
                        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
                            info = ydl.extract_info(url, download=False)

                        # Add to playlist
                        song = {}
                        song['webpage_url'] = info['webpage_url']
                        song['duration'] = info['duration']
                        song['title'] = info['title']
                        song['author'] = self.message.author
                        playlist.append(song)


                        # Construct add message
                        d = str(datetime.timedelta(seconds=int(song['duration'])))
                        out = bold("Added: [%s] %s" % (d, song['title']))

                        await self.client.send_message(bind_channel, out)

                    except:
                        # Unsupported URL
                        out = "Unsupported URL, see %s" % noembed(YDL_SITES)
                        raise CommandFailure(out)

            except HttpError as e:
                print('An HTTP error %d occurred:\n%s' \
                        % (e.resp.status, e.content))
                return "Something went wrong!"

        else:
            # Not a URL, search youtube using yt API
            try:
                playlist.append(youtube_search(args, self.message.author))
            except HttpError as e:
                print('An HTTP error %d occurred:\n%s' \
                        % (e.resp.status, e.content))
                return "Invalid link! (or something else went wrong :/)"

        # Nothing is playing, start the music event loop
        if not player or player.is_done():
            await music(voice, self.client, self.message.channel)


class Pause(M):
    desc = "Pauses music"

    async def eval(self):
        global bind_channel
        global paused
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()
        paused = True

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Paused Playing: [%s] %s" % (duration, player.title))

        # Change presence
        presence = PAUSE_UTF + player.title
        await self.client.change_presence(game=Game(name=presence))

        return out


class Resume(M):
    desc = "Resumes music"

    async def eval(self):
        global bind_channel
        global paused
        global player

        if not player:
            raise CommandFailure("Not playing anything!")

        player.resume()
        paused = False

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Resumed Playing: [%s] %s" % (duration, player.title))

        # Change presence
        presence = PLAY_UTF + player.title
        await self.client.change_presence(game=Game(name=presence))

        return out


class Skip(M):
    desc = "Skips a song. Defaults to the current song."

    def eval(self, pos=0):
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

        if 0 < pos <= len(playlist):
            # Remove the item from the playlist
            song = playlist.pop(pos-1)

            # Construct out message
            duration = str(datetime.timedelta(seconds=int(song['duration'])))
            out = bold("Removed: [%s] %s" % (duration, song['title']))

        elif pos == 0:
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

    async def eval(self):
        global bind_channel
        global paused
        global playlist

        if not player or player.is_done():
            raise CommandFailure("Not playing anything!")

        duration = str(datetime.timedelta(seconds=int(player.duration)))

        # Construct embed
        col = Colour.red() if paused else Colour.green()
        state = "Paused" if paused else "Playing"
        state = "Now " + state + ": [%s] %s" % (duration, player.title) 

        ti = "Up Next:" if len(playlist) > 0 else ""
        embed = Embed(title=ti, colour=col)
        embed.set_author(name=state)
        embed.set_footer(text="!m play [link/search]")
                        
        i = 0
        for song in playlist:
            i += 1

            duration = datetime.timedelta(seconds=int(song['duration']))
            title = "%d. [%s] %s" % (i, str(duration), song['title'])

            embed.add_field(name=title, 
                            value="Added by: %s" % nick(song['author']), 
                            inline=False)

        await self.client.send_message(bind_channel, embed=embed)


class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."

    async def eval(self):
        return await List.eval(self)


class Shuffle(M):
    desc = "Shuffles the playlist"

    def eval(self):
        global playlist

        if playlist and len(playlist) > 0:
            random.shuffle(playlist)
        else:
            raise CommandFailure("No playlist!")

        return "Shuffled Playlist!"


class Sh(M):
    desc = "See !m shuffle"

    def eval(self):
        return Shuffle.eval(self)


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

    while True:

        # Poll when there is no player or the player is finished
        if not player or player.is_done():

            try:
                # Play the next song in the list
                song = playlist.pop(0)
                url = song['webpage_url']

            except IndexError:
                # Nothing in playlist
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
            out = bold("Now playing:") + " [%s] %s" % (duration, presence)
            await client.send_message(bind_channel, out)

            # Change presence to the currently playing song
            presence = PLAY_UTF + presence
            await client.change_presence(game=Game(name=presence))

        await asyncio.sleep(SLEEP_INTERVAL)


def video_info(url, author):
    # This is the only function that gets the video info
    youtube = build(YOUTUBE_API_SERVICE_NAME, YOUTUBE_API_VERSION, 
          developerKey=DEVELOPER_KEY)

    # Split the url to get video id
    if url.startswith("http"):
        vid = url.split("v=")[1]
        vid = vid.split("&")[0]
    else:
        vid = url

    # API call
    videos = youtube.videos().list(
            part='snippet, contentDetails',
            id=vid
            ).execute()

    # Using vid id, will always return one item
    vid = videos['items'][0]

    # Construct info dict and return it
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

    # Split the url to get list id
    if url.startswith("http"):
        vid = url.split("list=")[1]
        vid = vid.split("&")[0]
    else:
        vid = url

    # API call
    videos = youtube.playlistItems().list(
            part='snippet, contentDetails',
            playlistId=vid,
            maxResults=MAX_RESULTS
            ).execute()

    # Get video metadata
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

    # Call the search.list method to retrieve results matching the specified
    # query term.
    search_response = youtube.search().list(
        q=query,
        part='id',
        maxResults=1,   # Only 1 video
        regionCode=GEO_REGION,
        type='video'
        ).execute()

    # Return the first video's info
    return video_info(search_response['items'][0]['id']['videoId'], author)
