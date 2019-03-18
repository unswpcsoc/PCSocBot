from commands.base import Command
from helpers import *
from discord import Embed
from discord import Color
import json
import random
from configstartup import config


EMOJI_FILE = config['FILES'].get('Emoji')
CHAR_LIMIT = 2000


class Emoji(Command):
    desc = "Prints a random custom emoji from the server"

    async def eval(self):
        emoji_list = self.server.emojis
        if not emoji_list:
            raise CommandFailure("Emoji list is empty!")
        return str(random.choice(emoji_list))


class Count(Emoji):
    desc = "Lists custom emojis and their use count.\n" + \
            "Use" + code("alpha") + " or " + code("count")

    async def eval(self, sort="count"):
        # Parse argument
        sort_i = 1
        if sort == "alpha":
            sort_i = 0
        elif sort != "count":
            raise CommandFailure("Please use " + code("alpha") + " or "
                                 + code("count"))

        # Open the JSON file or create a new dict to load
        try:
            with open(EMOJI_FILE, 'r') as old:
                emoji_dict = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('Emoji list is empty!')

        # Construct embed
        author = self.message.author
        embed = Embed(
                        title="Emoji Use Count:",
                        colour=author.colour,
                        timestamp=self.message.timestamp
                     )


        # print sorted list of emojis based on argument
        out = []
        for e, c in sorted(emoji_dict.items(), key=lambda elm: elm[sort_i]):
            out.append(f"{e} = {c}")

        length = len(out)
        embed.add_field(name="1", value="\n".join(out[:length]), inline=True)
        embed.add_field(name="2", value="\n".join(out[length:]), inline=True)

        return out


async def emojistats(message):
    # doesn't count if the message author is a bot
    if message.author.bot:
        return

    emoji_list = message.channel.server.emojis
    try:
        with open(EMOJI_FILE, 'r') as old:
            emoji_dict = json.load(old)
    except FileNotFoundError:
        emoji_dict = {}

    for emoji in emoji_list:
        if str(emoji) in message.content:
            emoji_dict[str(emoji)] = emoji_dict.get(str(emoji), 0) + 1

    with open(EMOJI_FILE, 'w') as new:
        json.dump(emoji_dict, new)
