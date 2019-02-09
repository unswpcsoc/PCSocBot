from commands.base import Command
from commands.state import *
from helpers import *

import asyncio

from discord import Game, Embed, Colour
import asyncio
import datetime
import isodate
import youtube_dl
import os
import random

DC_TIMEOUT = 300
PLIST_PREFIX = "https://www.youtube.com/playlist?list="
SLEEP_INTERVAL = 5
WRONG_PLIST = "https://www.youtube.com/watch?list="
VID_PREFIX = "https://www.youtube.com/watch?v="
YDL_SITES = "https://rg3.github.io/youtube-dl/supportedsites.html"


class M(Command):
    desc = "Music"
    channels_required = []


class Auto(M):
    desc = "Toggles autoplay."

    async def eval(self):
        check_bot_join(self.client, self.message)

        out = "Autoplay is now "
        out += bold(State.instance.toggleAuto())
        await State.instance.message(self.client, out)


class Add(Auto):
    desc = "Adds the autoplay suggestion for a playlist index. Defaults" +\
           " to the last item."

    def eval(self, index=-1):
        list_url = State.instance.getSong(index)['webpage_url']
        # Expensive call, use mp
        mp_call(auto_info, list_url, self.message.author)


class Reset(Auto):
    desc = "Reset the http session for autoplay i.e. cleans what Youtube" +\
           " has seen from autosuggestion requests"

    async def eval(self):
        out = State.instance.resetSession()
        await State.instance.message(self.client, out)


class List(M):
    desc = "Lists the playlist."

    async def eval(self):
        if not State.instance.isPlaying():
            raise CommandFailure("Not playing anything!")

        status, colour, title, footer = State.instance.stat()
        footer += " | `!m play [link/search]` or `!m auto add [index]`"
        embed = Embed(title=title, colour=colour)
        embed.set_author(name=status)
        embed.set_footer(text=footer)
        embed.set_thumbnail(url=State.instance.getSong()['thumb'])

        i = 0
        for song in State.instance.getPlaylist()[1:]:
            i += 1
            title = "%d. [%s] %s" % \
                (i, duration(song['duration']), song['title'])
            embed.add_field(name=title,
                            value="[Added by: %s](%s)"
                            % (nick(song['author']), song['webpage_url']),
                            inline=False)

        await State.instance.embed(self.client, embed)


class Ls(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."

    async def eval(self): return await List.eval(self)


class Np(M):
    desc = "See " + bold(code("!m") + " " + code("list")) + "."

    async def eval(self): return await List.eval(self)


class ListLimit(M):
    desc = "Sets playlist fetch limit. Mods only."
    roles_required = ["mod", "exec"]

    def eval(self, limit):
        try:
            limit = int(limit)
        except ValueError:
            raise CommandFailure("Please enter a valid integer")
        return State.instance.setLimit(limit)


class Pause(M):
    desc = "Pauses music."

    async def eval(self):
        out = State.instance.pause()
        await State.instance.updatePresence(self.client)
        return out


class Add(M):
    desc = "See " + bold(code("!m") + " " + code("play")) + "."

    async def eval(self, *args): return await Play.eval(self, *args)


class Play(M):
    desc = "Plays music. Binds commands to the channel invoked. Accepts YouTube"
    desc += " links, playlists (first %d entries), and more!\n" \
            % State.instance.getListLimit()
    desc += "Accepted site links: " + noembed(YDL_SITES)

    async def eval(self, *args):
        if len(args) == 0:
            raise CommandFailure("Invalid usage of command. Usage:\n" +
                                 self.tag_markup)

        args = " ".join(args)

        _, out = await State.instance.joinVoice(self.client, self.message)
        if out:
            await State.instance.message(self.client, out)
        M.channels_required.append(State.instance.getChannel())

        if args.startswith("http"):
            url = args

            if url.startswith(PLIST_PREFIX):
                # Expensive call, use mp
                mp_call(playlist_info, url, self.message.author)

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
                try:  # Not a youtube link, use youtube_dl
                    with youtube_dl.YoutubeDL() as ydl:
                        info = ydl.extract_info(url, download=False)
                    song = {}
                    song['webpage_url'] = info['webpage_url']
                    song['duration'] = info['duration']
                    song['title'] = info['title']
                    song['thumb'] = info['thumbnail']
                    song['author'] = self.message.author
                    out = State.instance.addSong(song)
                    await State.instance.message(self.client, out)
                except:
                    # Unsupported URL
                    out = "Unsupported URL, see %s" % noembed(YDL_SITES)
                    raise CommandFailure(out)

        else:  # Not a URL, search YouTube
            song = youtube_search(args, self.message.author)
            out = State.instance.addSong(song)
            await State.instance.message(self.client, out)

        # Music not running, start it
        if not State.instance.running:
            State.instance.running = True
            await music(State.instance.getVoice(), self.client,
                        self.message.channel)


class Remove(M):
    desc = "Removes a song from the playlist. Defaults to the current song."

    def eval(self, index=0):
        check_bot_join(self.client, self.message)
        return State.instance.remove(index)


class Repeat(M):
    desc = "Toggle repeat for the current song or the whole playlist. "
    desc += "Accepted arguments are: `none`, `song` and `list`."

    async def eval(self, mode):
        check_bot_join(self.client, self.message)
        out = State.instance.repeat(mode)
        await State.instance.updatePresence(self.client)
        return bold("Repeat mode set to: %s" % out)


class Resume(M):
    desc = "Resumes music."

    async def eval(self):
        out = State.instance.resume()
        await State.instance.updatePresence(self.client)
        return out


class Rm(M):
    desc = "See " + bold(code("!m") + " " + code("remove")) + "."

    def eval(self, pos=0): return Remove.eval(self, pos)


class Rp(M):
    desc = "See " + bold(code("!m") + " " + code("repeat")) + "."

    def eval(self, mode): return Repeat.eval(self.mode, mode)


class Shuffle(M):
    desc = "Shuffles the playlist in place."

    def eval(self):
        check_bot_join(self.client, self.message)
        return State.instance.shuffle()


class Sh(M):
    desc = "See " + bold(code("!m") + " " + code("shuffle")) + "."

    def eval(self): return Shuffle.eval(self)


class Skip(M):
    desc = "Skips the current song."

    def eval(self):
        check_bot_join(self.client, self.message)
        if not State.instance.isPlaying():
            raise CommandFailure("Not playing anything!")

        State.instance.stop()
        out = bold("Skipped:") + " [%s] %s" % \
            (State.instance.playerDuration(), State.instance.playerTitle())
        return out


class Stop(M):
    desc = "Stops playing but persists in voice. Also turns auto off."

    async def eval(self):
        if not State.instance.isPlaying():
            return "Not playing anything!"
        await State.instance.clean(self.client)
        State.instance.setAuto(False)
        out = bold("Stopped Playing")
        await State.instance.message(self.client, out)


class Volume(M):
    desc = "Volume adjustment. Mods only."
    roles_required = ["mod", "exec"]

    def eval(self, *level):
        return State.instance.volume(*level)


class V(M):
    desc = "See " + bold(code("!m") + " " + code("volume")) + "."
    roles_required = ["mod", "exec"]

    def eval(self, *level): return Volume.eval(self, *level)


async def music(voice, client, channel):
    try:
        dc_ticker = 0
        paused_dc = False
        was_playing = False

        # Begin http session for better autoplay suggestions
        State.instance.beginSession()
        while True:
            """ Multiprocessing notes
            - Can't run coroutines in different processes using multiprocessing
            - New solution: run expensive synchronous actions (auto_get, 
            playlist_info, etc.) in new process and read from queue in every 
            event loop iteration
            """
            await asyncio.sleep(SLEEP_INTERVAL)
            # MUSIC
            # Poll multiprocessing queue
            song = State.instance.qGet()
            if song is not None:
                out = State.instance.addSong(song)
                await State.instance.message(client, out)
                State.instance.freeLock()
                was_playing = False
            # Need to softlock auto-adding from handlePop
            # DO NOT MESS WITH THE LOCK OR IT WILL MESS WITH YOU
            if State.instance.isLocked():
                continue
            # Handle player done
            if State.instance.isDone():
                if State.instance.isListEmpty() and not State.instance.isAuto():
                    was_playing = False
                    continue
                if not State.instance.isListEmpty():
                    if was_playing:
                        out = State.instance.handlePop(client)
                        if State.instance.isLocked():
                            continue
                        if out:
                            await State.instance.message(client, out)
                    was_playing = True
                    out = await State.instance.playNext()
                    if out == None:
                        continue
                    await State.instance.message(client, out)
                    await State.instance.updatePresence(client)
            # CISUM
            # AUDIENCE
            # Handle no audience
            if len(voice.channel.voice_members) <= 1:
                if dc_ticker >= DC_TIMEOUT:
                    out = "Timeout of [%s] reached," % duration(DC_TIMEOUT)
                    out += " Disconnecting from %s," % code(voice.channel.name)
                    out += " Unbinding from %s" % \
                        chan(State.instance.getChannel().id)
                    await State.instance.message(client, out)
                    break
                # Start counting
                if not paused_dc:
                    paused_dc = True
                    State.instance.pause()
                    name = voice.channel.name
                    out = bold("Nobody listening in %s, Pausing" % code(name))
                    await State.instance.message(client, out)
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
            # ECNEIDUA
    finally:
        await voice.disconnect()
        M.channels_required.clear()
        await State.instance.clean(client)
        State.instance.running = False
        State.instance.cleanSession()
        return
