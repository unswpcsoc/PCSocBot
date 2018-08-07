from commands.base import Command
from googletrans import Translator
from discord import Embed
from helpers import *

import json

HISTORY = 10
HISTORY_LIMIT = 100
FLAGS_FILE = "files/flags.json"
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
IMAGE_EXTENSIONS = ['png', 'jpg', 'jpeg', 'gif']

class Translate(Command):
    desc = "Translates a marked message to or from Korean."

    async def eval(self, index=0):
            if int(index) < 0:
                raise CommandFailure('Invalid index!')

            archive = await create_archive(self.client.logs_from(self.message.channel,
                                                                    limit=HISTORY_LIMIT))

            if not archive:
                raise CommandFailure('No tagged messages in the last %s found!' % HISTORY_LIMIT)

            if int(index) >= len(archive):
                raise CommandFailure('Invalid index!')

            lang_list = archive[int(index)]
            for entry in lang_list:
                #translate contents
                translator = Translator()
                entry.content = translator.translate(entry.content, dest=entry.language).text

                channel = self.message.channel
                header = ''
                footer = "Translation to %s requested by %s" % (entry.flag, self.name)
                await self.client.send_message(channel, header, embed=entry.as_embed(footer))


class Entry():
    def __init__(self, message, language, flag):
        self.author = message.author
        self.content = message.clean_content
        self.timestamp = message.timestamp
        self.language = language
        self.flag = flag


    def as_embed(self, title):
        embed = Embed(description=self.content,
                     colour=self.author.colour,
                     timestamp=self.timestamp)
        embed.set_author(name=nick(self.author),
                         icon_url=AVATAR_FORMAT.format(self.author))
        embed.set_footer(text=title)
        return embed


async def create_archive(logs):
    i = 0
    archive = []

    # load dict of flags
    try:
        with open(FLAGS_FILE, 'r') as f:
            flags = json.load(f)
    except FileNotFoundError:
        raise CommandFailure('Flags file does not exist!')

    # Iterate through message history
    async for message in logs:
        # Break when we've found HISTORY messages
        if i >= HISTORY:
            break

        reactions = [x.emoji for x in message.reactions]
        lang_list = []
        for r in reactions:
            if r in flags:
                lang_list.append(Entry(message, flags[r], r))
                i += 1

        if lang_list:
            archive.append(lang_list)
                
    return archive

    