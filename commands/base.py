import asyncio
from collections import OrderedDict
import inspect

import discord
from pony.orm import db_session
from mutagen.mp3 import MP3

from helpers import classproperty, code, bold, underline, CommandFailure

PREFIX = '!'


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
    db_required = False
    desc = bold('PCSocBot') + ' - PC Enthusiasts society Discord bot made with discord.py by Matt Stark'
    pprint = {}

    @classproperty
    def name(cls):
        return cls.__name__.lower()

    def __init__(self, client, message):
        self.client = client
        self.message = message
        self.user = message.author.id
        self.name = message.author.name
        self.members = message.server.members

    async def init(self, *args):
        argspec = inspect.getargspec(self.eval)
        if len(argspec.args) == len(args) + 1 or argspec.varargs:
            try:
                self.check_permissions()
                self.output = await self.eval(*args) if inspect.iscoroutinefunction(self.eval) else self.eval(*args)
            except CommandFailure as e:
                self.output = e.args[0]
        elif self.tag_prefix_list:
            self.output = "Invalid usage of command. Usage:\n" + self.tag_markup
        else:
            self.output = "Invalid command\n" + self.help
        return self.output

    def eval(self):
        return self.help

    @classproperty
    def tag_prefix_list(cls):
        if cls.__base__ == object:
            return []
        return cls.__base__.tag_prefix_list + [cls.name]

    @classproperty
    def tag_markup(cls):
        func_args = inspect.getargspec(cls.eval).args[1:] + [inspect.getargspec(cls.eval).varargs]
        if func_args[-1] is None: func_args.pop()
        prefix = cls.tag_prefix_list
        prefix[0] = PREFIX + prefix[0]
        return ' '.join(bold(code(item)) for item in prefix) + ' ' + \
               ' '.join(underline(code(cls.pprint.get(item, item))) for item in func_args)

    @classproperty
    def help(cls):
        if cls.subcommands:
            lines = [cls.desc, '', bold('Commands' if cls.__base__ == object else 'Subcommands')]
            if cls.__base__ != object:
                lines = [cls.tag_markup] + lines
            for command in cls.subcommands.values():
                lines.append(command.tag_markup)
                lines.append(command.desc)
        else:
            lines = [cls.tag_markup , cls.desc]
        return '\n'.join(lines)

    def get_member(self, id):
        return discord.utils.get(self.members, id=str(id))

    def check_permissions(self):
        if self.roles_required is not None:
            for role in self.message.author.roles:
                if role.name.lower() in self.roles_required:
                    return
            raise CommandFailure("You need to be a %s to use that command" % \
                                 " or ".join(self.roles_required))


    audio_playing = False
    async def play_audio(self, file):
        channel = self.message.author.voice_channel
        if channel is None:
            raise CommandFailure("You need to join a voice channel to use this command")
        if Command.audio_playing:
            raise CommandFailure("Can only play one audio file at once")
        Command.audio_playing = True  # coroutines, no mutex required
        voice = await self.client.join_voice_channel(channel)
        player = voice.create_ffmpeg_player('files/' + file)

        player.start()
        duration = MP3('files/' + file).info.length
        await asyncio.sleep(duration)
        await voice.disconnect()
        Command.audio_playing = False
