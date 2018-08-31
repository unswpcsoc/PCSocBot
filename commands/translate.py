from commands.base import Command
from discord import Embed
from helpers import *
from utils.embed_table import EmbedTable

import googletrans
import json
import re

HISTORY = 10
HISTORY_LIMIT = 100
FLAGS_FILE = "files/flags.json"
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
TRANSLATE_COLOR = int('4885ed', 16)

class Translate(Command):
    desc = """Translates a marked message to another language (removes all emojis).
    React with a flag and type `!translate` to translate the message to the corresponding language.
    Optional `index` argument defines offset of message to translate in list of marked messages."""

    async def eval(self, index=0):
        if int(index) < 0:
            raise CommandFailure('Invalid index!')

        archive = await create_list(self.client.logs_from(self.message.channel,
                                                            limit=HISTORY_LIMIT))

        if not archive:
            raise CommandFailure('No tagged messages in the last %s found!' % HISTORY_LIMIT)

        if int(index) >= len(archive):
            raise CommandFailure('Invalid index!')

        lang_list = archive[int(index)]
        for entry in lang_list:
            #translate contents
            translator = googletrans.Translator()
            text = removeEmojis(entry.content)
            entry.content = translator.translate(text, dest=entry.language).text

            channel = self.message.channel
            header = ''
            footer = "Translation to %s requested by %s" % (entry.flag, self.name)
            await self.client.send_message(channel, header, embed=entry.as_embed(footer))


class Add(Translate):
    desc = "Adds mapping between an emoji and a Google API language code. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, emoji, langcode):
        #check if valid language code
        if langcode not in googletrans.LANGUAGES:
            raise CommandFailure("Language code not valid!")

        # Open the JSON file or create a new dict to load
        try:
            with open(FLAGS_FILE, 'r') as old:
                flags = json.load(old)
        except FileNotFoundError:
            flags = {}

        flags[emoji] = langcode

        # Write the formats to the JSON file
        with open(FLAGS_FILE, 'w') as new:
            json.dump(flags, new)

        return "%s now maps to %s!" % (emoji, code(langcode))


class Remove(Translate):
    desc = "Removes a emoji->langcode mapping if it exists. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, emoji):
        # Open the JSON file or create a new dict to load
        try:
            with open(FLAGS_FILE, 'r') as old:
                flags = json.load(old)
            langcode = flags.pop(emoji)

        except (FileNotFoundError, KeyError, ValueError):
            raise CommandFailure("Emoji %s not found!" % emoji)

        # Write the formats to the JSON file
        with open(FLAGS_FILE, 'w') as new:
            json.dump(flags, new)

        return "Mapping of %s to %s removed!" % (emoji, code(langcode))

class List(Translate):
    desc = "Lists all a emoji->langcode mappings. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self):
        # Open the JSON file, skip if it does not exist
        try:
            with open(FLAGS_FILE, 'r') as old:
                flags = json.load(old)
        except FileNotFoundError:
            raise CommandFailure("Emoji list is empty!")

        return EmbedTable(fields=['Emoji->Code'], 
                         table=[(emoji+'->'+code,) for emoji, code in flags.items()], 
                         colour=TRANSLATE_COLOR)


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


async def create_list(logs):
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


def removeEmojis(text):
    emoji_pattern = re.compile("["
            u"\U0001F600-\U0001F64F"  # emoticons
            u"\U0001F300-\U0001F5FF"  # symbols & pictographs
            u"\U0001F680-\U0001F6FF"  # transport & map symbols
            u"\U0001F1E0-\U0001F1FF"  # flags (iOS)
                            "]+", flags=re.UNICODE)

    return emoji_pattern.sub(r'', text)