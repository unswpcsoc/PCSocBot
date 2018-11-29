from commands.base import Command
from commands.state import *
from helpers import *

from discord import Game, Embed
import asyncio, youtube_dl, os

DC_TIMEOUT = 300
PLIST_PREFIX = "https://www.youtube.com/playlist?list="
SLEEP_INTERVAL = 1
WRONG_PLIST = "https://www.youtube.com/watch?list="
VID_PREFIX = "https://www.youtube.com/watch?v="
YDL_SITES = "https://rg3.github.io/youtube-dl/supportedsites.html"

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
            to the last item."

    def eval(self, index=-1):
        url = auto_get(State.instance.getSong(index)['webpage_url'])
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
        for song in State.instance.getPlaylist()[1:]:
            i += 1
            title = "%d. [%s] %s" % \
                    (i, duration(song['duration']), song['title'])
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
        try: limit = int(limit)
        except ValueError: raise CommandFailure("Please enter a valid integer")
        return State.instance.setLimit(limit)

class Pause(M):
    desc = "Pauses music"

    async def eval(self):
        out = State.instance.pause()
        await State.instance.updatePresence(self.client)
        return out

class Play(M):
    desc = "Plays music. Binds commands to the channel invoked.\n"
    desc += "Accepts youtube links, playlists (first %d entries), and more!\n"
    desc += noembed(YDL_SITES)

    async def eval(self, *args):
        args = " ".join(args)

        voice, out = await State.instance.joinVoice(self.client, self.message)
        if out: await State.instance.message(self.client, out)
        M.channels_required.append(State.instance.getChannel())

        if args.startswith("http"):
            url = args

            if url.startswith(PLIST_PREFIX):
                songs = playlist_info(url, self.message.author)
                out = State.instance.addList(songs)
                await State.instance.message(self.client, out)

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
                try: # Not a youtube link, use youtube_dl
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

        else: # Not a URL, search YouTube
            song = youtube_search(args, self.message.author)
            out = State.instance.addSong(song)
            await State.instance.message(self.client, out)

        # Music not running, start it
        if not State.instance.running:
            State.instance.running = True
            await music(State.instance.getVoice(), self.client, \
                    self.message.channel)

class Remove(M):
    desc = "Removes a song from the playlist. Defaults to the current song"

    def eval(self, index=0):
        check_bot_join(self.client, self.message)
        return State.instance.remove(index)

class Repeat(M):
    desc = "Toggle repeat for the current song or the whole playlist. "
    desc += "Accepted arguments are: 'none', `song` and `list`."

    async def eval(self, mode):
        check_bot_join(self.client, self.message)
        out = State.instance.repeat(mode)
        await State.instance.updatePresence(self.client)
        return bold("Repeat mode set to: %s" % out)

class Resume(M):
    desc = "Resumes music"

    async def eval(self):
        out = State.instance.resume()
        await State.instance.updatePresence(self.client)
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
        check_bot_join(self.client, self.message)
        return State.instance.shuffle()

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
    desc = "Stops playing but persists in voice. Also turns auto off."

    async def eval(self):
        if not State.instance.isPlaying():
            return "Not playing anything!"
        await State.instance.clean()
        State.instance.setAuto(False)

class Volume(M):
    desc = "Volume adjustment. Mods only."
    roles_required = [ "mod", "exec" ]

    def eval(self, level):
        return State.instance.volume(level)

class V(M):
    desc = "See " + bold(code("!m") + " " + code("volume")) + "."
    roles_required = [ "mod", "exec" ]

    def eval(self, level): return Volume.eval(self, level)

async def music(voice, client, channel):
    dc_ticker = 0
    paused_dc = False
    was_playing = False

    while True:
        # Handle player done
        if State.instance.isDone():
            if State.instance.isListEmpty() and State.instance.hasPlayer():
                await State.instance.clean()
                was_playing = False
                out = bold("Stopped Playing")
                await State.instance.message(client, out)

            if not State.instance.isListEmpty():
                if was_playing: 
                    out = State.instance.handlePop(client)
                    if out: await State.instance.message(client, out)
                was_playing = True
                out = await State.instance.playNext()
                if out == None: continue
                await State.instance.message(client, out)
                await State.instance.updatePresence(client)

        # Handle no audience
        if len(voice.channel.voice_members) <= 1:   

            if dc_ticker >= DC_TIMEOUT: 
                await State.instance.clean()
                await voice.disconnect()
                M.channels_required.clear()
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

    State.instance.running = False
    return
