import asyncio
import os
from collections import OrderedDict
import inspect

import discord
from pony.orm import db_session
from mutagen.mp3 import MP3

from helpers import *
from configstartup import config

PREFIX = '~' if os.environ.get('DEBUG') else '!'
CHAR_LIM = 2000

player = None


class Tree(type):
    def __init__(cls, name, bases, clsdict):
        assert len(bases) < 2  # no multiple inheritance for commands
        if bases:
            bases[0].subcommands[cls.name] = cls
        super().__init__(name, bases, clsdict)

        cls.subcommands = OrderedDict()
        if cls.db_required:
            cls.eval = db_session(cls.eval)


class Command(metaclass=Tree):
    roles_required = None
    channels_required = None
    db_required = False
    disabled = False
    desc = 'The UNSW Computer Enthusiasts Society Discord Bot\n'
    desc += noembed('https://github.com/unswpcsoc/PCSocBot')
    pprint = dict()
    EMBED_COLOR = int('f0cf20', 16)

    @classproperty
    def name(cls):
        return cls.__name__.lower()

    def __init__(self, client, message):
        self.client = client
        self.message = message
        self.user = message.author.id
        self.name = message.author.name
        self.server = message.server
        self.members = message.server.members

    async def init(self, *args):
        argspec = inspect.getargspec(self.eval)
        if len(argspec.args) == len(args) + 1 or argspec.varargs or argspec.defaults:
            try:
                self.check_permissions()
                self.check_channels()
                return await self.eval(*args) if inspect.iscoroutinefunction(self.eval) else self.eval(*args)
            except CommandFailure as e:
                return e.args[0]
        elif self.tag_prefix_list:
            return "Invalid usage of command. Usage:\n" + self.tag_markup

    def eval(self):
        return self.help

    @classproperty
    def tag_prefix_list(cls):
        if cls.__base__ == object:
            return []
        return cls.__base__.tag_prefix_list + [cls.name]

    @classproperty
    def tag_markup(cls):
        func_args = inspect.getargspec(
            cls.eval).args[1:] + [inspect.getargspec(cls.eval).varargs]
        if func_args[-1] is None:
            func_args.pop()
        prefix = cls.tag_prefix_list
        prefix[0] = PREFIX + prefix[0]
        return ' '.join(bold(code(item)) for item in prefix) + ' ' + \
               ' '.join(underline(code(cls.pprint.get(item, item)))
                        for item in func_args)

    @classproperty
    def help(cls):
        out = []
        if cls.disabled:
            # Don't give any help message for a disabled command
            return
        if cls.subcommands:
            cmd = bold('Commands' if cls.__base__ == object else 'Subcommands')
            lines = [cls.desc, '', cmd]
            if cls.__base__ != object:
                lines = [cls.tag_markup] + lines
            for command in cls.subcommands.values():
                if command.disabled:
                    # Don't include disabled commands in helpme message
                    continue
                lines.append(command.tag_markup)
                lines.append(command.desc)
                if len('\n'.join(lines)) > CHAR_LIM:
                    out.append('\n'.join(lines[:-2]))
                    lines = [cmd + ' continued...'] + lines[-2:]
            out.append('\n'.join(lines))
        else:
            lines = [cls.tag_markup, cls.desc]
            out = ['\n'.join(lines)]
        h = '\n\nType ' + bold(code('!h(elpme) [<command/subcommand>]')) + \
            'for more info on a command'
        if len(out[-1] + h) > CHAR_LIM:
            out.append(h)
        else:
            out[-1] += h
        return out

    def from_id(self, id):
        return discord.utils.get(self.members, id=str(id))

    def from_name(self, name):
        return discord.utils.find(lambda x: x.name.lower() == name.lower(), self.members)

    def check_permissions(self):
        if self.roles_required:
            role_ids = [config['ROLES'].get(r.lower())
                        for r in self.roles_required]
            if any(role.id in role_ids for role in self.message.author.roles):
                # User has permissions
                return
            raise CommandFailure("You need to be a %s to use that command" %
                                 " or ".join(self.roles_required))

    def check_channels(self):
        cr = self.channels_required
        if cr is not None and len(cr) > 0 and self.message.channel not in cr:
            raise CommandFailure("You need to use this command in %s" %
                                 " or ".join([chan(x.id) for x in cr]))

    async def play_mp3(self, file, volume, quiet=False):
        global player

        channel = self.message.author.voice.voice_channel

        if not channel:
            if not quiet:
                raise CommandFailure(
                    "You need to join a voice channel to use this command")
            else:
                return

        if player:
            if not quiet:
                raise CommandFailure("Already playing something!")
            else:
                return

        # Check if bot is connected already in the server
        vclients = list(self.client.voice_clients)
        voices = [x.server for x in vclients]
        try:
            # Get the voice channel
            v_index = voices.index(self.message.server)
            voice = vclients[v_index]
        except ValueError:
            # Not connected, join a vc
            voice = await self.client.join_voice_channel(channel)

        # Get voice event loop
        #loop = voice.loop
        player = voice.create_ffmpeg_player(
            'files/' + file)  # , after=lambda: \
        #asyncio.run_coroutine_threadsafe(voice_client.disconnect(), loop)
        player.volume = volume/100
        player.start()

        duration = MP3('files/' + file).info.length
        await asyncio.sleep(duration)
        await voice.disconnect()

        # Reset player
        player = None
