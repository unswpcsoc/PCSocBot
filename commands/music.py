from commands.base import Command
from commands.playing import CURRENT_PRESENCE
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
DC_TIMEOUT = 60

bind_channel = None
paused = False
player = None
playlist = []
repeat = "none"
presence = CURRENT_PRESENCE
volume = float(12)
list_limit = 10


class M(Command):
    desc = "Music"
    channels_required = []


class Play(M):
    desc = "Plays music. Binds commands to the channel invoked.\n"
    desc += "Accepts youtube links, playlists (first %d entries), and more!\n"
    desc += noembed(YDL_SITES) + "\n"
    desc += "Note: Playlists fetching through the YT API is limited to 50 vids"

    async def eval(self, *args):
        global bind_channel
        global player
        global playlist

        args = " ".join(args)

        # Check if user is connected to a vc
        channel = self.message.author.voice.voice_channel
        if not channel:
            raise CommandFailure("Please join a voice channel first")

        # Check if bot is joined
        # Note: don't use check_bot_join()
        vclients = list(self.client.voice_clients)
        voices = [ x.server for x in vclients ]
        try:
            # Get the voice channel
            v_index = voices.index(self.message.server)
            voice = vclients[v_index]
            was_connected = True
        except ValueError:
            # Not connected, join a vc
            voice = await do_join(self.client, self.message)
            was_connected = False
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
                        out += bold("[%s] %s" % (str(d), song['title']))
                        out += "\n"

                    await self.client.send_message(bind_channel, out)

                elif url.startswith(WRONG_PLIST):
                    # Incorrect truncation, suggest the correct one
                    out = "Warning: invalid Playlist URL\n"
                    out += "Please change `watch` to `playlist`"

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
                print('%s YOUTUBE: A HTTP error %d occurred:\n%s' \
                        % (timestamp(), e.resp.status, e.content))
                return "Invalid link! (or something else went wrong :/)"

        else:
            # Not a URL, search youtube using yt API
            try:
                song = youtube_search(args, self.message.author)
                playlist.append(song)

                # Construct add message
                d = str(datetime.timedelta(seconds=int(song['duration'])))
                out = bold("Added: [%s] %s" % (d, song['title']))

                await self.client.send_message(bind_channel, out)

            except HttpError as e:
                print('%s YOUTUBE: A HTTP error %d occurred:\n%s' \
                        % (timestamp(), e.resp.status, e.content))
                return "Invalid link! (or something else went wrong :/)"

        # Nothing is playing and we weren't in vc, start the music event loop
        if (not player or player.is_done()) and not was_connected:
            await music(voice, self.client, self.message.channel)


class Pause(M):
    desc = "Pauses music"

    async def eval(self):
        global bind_channel
        global paused
        global player
        global presence
        global repeat

        if not player:
            raise CommandFailure("Not playing anything!")

        player.pause()
        paused = True

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Paused Playing: [%s] %s" % (duration, player.title))

        # Change presence
        presence = PAUSE_UTF + player.title
        if repeat == "song": presence = REPEAT_SONG_UTF + presence
        if repeat == "list": presence = REPEAT_LIST_UTF + presence

        await self.client.change_presence(game=Game(name=presence))

        return out


class Resume(M):
    desc = "Resumes music"

    async def eval(self):
        global bind_channel
        global paused
        global player
        global presence
        global repeat

        if not player:
            raise CommandFailure("Not playing anything!")

        player.resume()
        paused = False

        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Resumed Playing: [%s] %s" % (duration, player.title))

        # Change presence
        presence = PLAY_UTF + player.title
        if repeat == "song": presence = REPEAT_SONG_UTF + presence
        if repeat == "list": presence = REPEAT_LIST_UTF + presence

        await self.client.change_presence(game=Game(name=presence))

        return out


class Skip(M):
    desc = "Skips the current song. Does not skip if repeat is `song`"

    def eval(self):
        # Check if connected to a voice channel
        check_bot_join(self.client, self.message)

        if not player: raise CommandFailure("Not playing anything!")
        if player.is_done(): raise CommandFailure("Not playing anything!")

        # Construct out message
        duration = str(datetime.timedelta(seconds=int(player.duration)))
        out = bold("Skipped: [%s] %s" % (duration, player.title))

        # Stop player, triggers music() to queue up the next song
        # according to the repeat policy
        player.stop()

        return out


class Remove(M):
    desc = "Removes a song from the playlist. Defaults to the current song"

    def eval(self, pos=0):
        global bind_channel
        global player
        global playlist
        global repeat

        try:
            pos = int(pos)
        except ValueError:
            raise CommandFailure("Not a valid position!")

        if not player:
            raise CommandFailure("Not playing anything!")

        if pos < 0 or pos > len(playlist):
            raise CommandFailure("Not a valid position!")

        # Check if connected to a voice channel
        check_bot_join(self.client, self.message)

        # Remove the item from the playlist
        song = playlist.pop(pos)

        # Construct out message
        duration = str(datetime.timedelta(seconds=int(song['duration'])))
        out = bold("Removed: [%s] %s" % (duration, song['title']))

        # Kill the player if we remove the currently playing song
        if pos == 0: player.stop()

        return out


class Rm(M):
    desc = "See " + bold(code("!m") + " " + code("remove")) + "."

    def eval(self, pos=0): return Remove.eval(self, pos)


class Stop(M):
    desc = "Stops playing but persists in voice"

    def eval(self):
        global bind_channel
        global player
        global playlist

        if not player:
            return "Not playing anything!"

        player.stop()
        playlist.clear()


class Volume(M):
    desc = "Volume adjustment. Mods only."
    roles_required = [ "mod", "exec" ]

    def eval(self, level):
        global bind_channel
        global player
        global volume

        if not player:
            raise CommandFailure("Not playing anything!")

        try:
            level = float(level)
        except ValueError:
            raise CommandFailure("Please enter a number between 0-100")

        if 0 <= level <= 100:
            # Change the global levelume
            volume = level
            player.volume = volume/100
            out = "Volume changed to %f%%" % level
            return out
        else:
            raise CommandFailure("Please enter a number from 0-100")


class V(M):
    desc = "See " + bold(code("!m") + " " + code("volume")) + "."
    roles_required = [ "mod", "exec" ]

    def eval(self, level): return Volume.eval(self, level)


class List(M):
    desc = "Lists the playlist."

    async def eval(self):
        global bind_channel
        global paused
        global playlist
        global repeat

        if not player or player.is_done():
            raise CommandFailure("Not playing anything!")

        duration = str(datetime.timedelta(seconds=int(player.duration)))

        # Construct embed
        col = Colour.red() if paused else Colour.green()
        state = "Paused" if paused else "Playing"
        state = "Now " + state + ": [%s] %s" % (duration, player.title) 

        # Construct title
        ti = "Up Next: (Repeat: %s)" % repeat if len(playlist) > 1 else ""

        embed = Embed(title=ti, colour=col)
        embed.set_author(name=state)
        embed.set_footer(text="!m play [link/search]")
                        
        # Get fields
        i = 0
        for song in playlist[1:]:
            i += 1

            duration = datetime.timedelta(seconds=int(song['duration']))
            title = "%d. [%s] %s" % (i, str(duration), song['title'])

            embed.add_field(name=title, 
                            value="Added by: %s" % nick(song['author']), 
                            inline=False)

        await self.client.send_message(bind_channel, embed=embed)


class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."

    async def eval(self): return await List.eval(self)


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


class Repeat(M):
    desc = "Toggle repeat for the current song or the whole playlist. "
    desc += "Accepted arguments are: 'none', `song` and `list`.\n"

    async def eval(self, mode):
        global repeat
        global presence

        # Check if connected to a voice channel
        check_bot_join(self.client, self.message)

        if mode.lower() == "song": 
            if repeat != "none": presence = presence[1:]
            presence = REPEAT_SONG_UTF + presence
            repeat = mode.lower()

        elif mode.lower() == "list": 
            if repeat != "none": presence = presence[1:]
            presence = REPEAT_LIST_UTF + presence
            repeat = mode.lower()

        elif mode.lower() == "none": 
            if repeat != "none": presence = presence[1:]
            repeat = mode.lower()

        else: 
            raise CommandFailure("Not a valid argument!")

        # Change presence
        await self.client.change_presence(game=Game(name=presence))

        return bold("Repeat mode set to: %s" % repeat)


class Rp(M):
    desc = "See " + bold(code("!m") + " " + code("repeat")) + "."

    def eval(self, mode): return Repeat.eval(self.mode)


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


# HELPER FUNCTIONS #


def check_bot_join(client, message):
    vclients = list(client.voice_clients)
    voices = [ x.server for x in vclients ]
    try:
        v_index = voices.index(message.server)
    except ValueError:
        # Bot is not connected to a voice channel in this server
        raise CommandFailure("Please `!join` a voice channel first") 


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
    await client.send_message(bind_channel, bold(out))

    # Set bitrate
    voice.encoder_options(sample_rate=SAMPLE_RATE, channels=2)

    return voice


async def music(voice, client, channel):
    global bind_channel
    global paused
    global player
    global playlist
    global presence
    global volume

    # Sentinel value for paused state
    # Could use player.is_playing() but this is faster
    paused_dc = False
    dc_timer = 0
    was_playing = False
    while True:

        # Poll for no player or the player is finished
        if not player or player.is_done():

            name = voice.channel.name

            if len(playlist) > 0:

                # Check for repeating modes
                if repeat == "song": pass
                elif repeat == "list": playlist.append(playlist.pop(0))
                else: 
                    if was_playing: playlist.pop(0)

                # Play the next song in the list
                if len(playlist) > 0: song = playlist[0]
                else: continue

                url = song['webpage_url']

                player = await voice.create_ytdl_player(url)
                player.start()
                player.volume = volume/100  # "That's how you get tinnitus!"

                was_playing = True

                # Print the message in the supplied channel
                duration = str(datetime.timedelta(seconds=int(song['duration'])))
                presence = song['title']
                out = bold("Now Playing: [%s] %s" % (duration, presence))
                await client.send_message(bind_channel, out)

                # Change presence to the currently playing song
                presence = PLAY_UTF + presence
                if repeat == "song": presence = REPEAT_SONG_UTF + presence
                if repeat == "list": presence = REPEAT_LIST_UTF + presence

                await client.change_presence(game=Game(name=presence))

            else:
                if player:
                    # Reset player
                    player = None

                    was_playing = False

                    out = bold("Stopped Playing")
                    await client.send_message(bind_channel, out)

                    # Change presence back
                    await client.change_presence(game=Game(\
                                                name=CURRENT_PRESENCE))

                # Poll for listeners
                if len(voice.channel.voice_members) <= 1:
                    dc_timer += 1
                    if dc_timer == DC_TIMEOUT:
                        await voice.disconnect()

                        d = str(datetime.timedelta(seconds=int(DC_TIMEOUT)))
                        out = "Timeout of [%s] reached," % d
                        out += " Disconnecting from %s," % code(name)
                        out += " Unbinding from %s" % chan(bind_channel.id)

                        # Flush channels required
                        M.channels_required.clear()

                        await client.send_message(bind_channel, bold(out))
                        break

                else: dc_timer = 0

        else:   # Something is playing

            # Poll for no listeners in channel
            if len(voice.channel.voice_members) <= 1:

                # while playing
                if player.is_playing():
                    player.pause()
                    paused_dc = True
                    paused = True

                    # Change presence
                    presence = PAUSE_UTF + player.title
                    if repeat == "song": presence = REPEAT_SONG_UTF + presence
                    if repeat == "list": presence = REPEAT_LIST_UTF + presence

                    await client.change_presence(game=Game(name=presence))

                    out = bold("Nobody listening in %s, Pausing" % code(name))
                    await client.send_message(bind_channel, out)

                elif paused_dc:    # Make sure we are the ones who paused
                    # Careful, if SLEEP_INTERVAL changes, the duration will change
                    dc_timer += 1

                if dc_timer >= DC_TIMEOUT: 
                    await voice.disconnect()

                    d = str(datetime.timedelta(seconds=int(DC_TIMEOUT)))
                    out = "Timeout of [%s] reached," % d
                    out += " Disconnecting from %s," % code(name)
                    out += " Unbinding from %s" % chan(bind_channel.id)

                    # Flush channels required
                    M.channels_required.clear()

                    await client.send_message(bind_channel, bold(out))

                    # Change presence back
                    await client.change_presence(game=Game(
                                                name=CURRENT_PRESENCE))

                    # Reset paused
                    paused = False
                    break

            # Poll for listeners when we have paused_dc ourselves
            # Since listeners can pause themselves, we must use our own value
            if len(voice.channel.voice_members) > 1 and paused_dc:
                player.resume()
                paused_dc = False
                paused = False
                dc_timer = 0

                # Change presence
                presence = PLAY_UTF + player.title
                if repeat == "song": presence = REPEAT_SONG_UTF + presence
                if repeat == "list": presence = REPEAT_LIST_UTF + presence

                await client.change_presence(game=Game(name=presence))

                out = bold("Somebody has joined %s! Resuming" % code(name))
                await client.send_message(bind_channel, out)

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
            maxResults=list_limit
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
    try:
        return video_info(search_response['items'][0]['id']['videoId'], author)
    except IndexError:
        raise CommandFailure(bold("Couldn't find %s" % query))
