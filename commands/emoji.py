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
    desc = "Lists custom emojis and their use count"

    async def eval(self):
        # Open the JSON file or create a new dict to load
        try:
            with open(EMOJI_FILE, 'r') as old:
                emoji_dict = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('Emoji list is empty!')

        # print list of emojis
        out = bold("Emoji use count:") + '\n'
        for e, c in emoji_dict.items():
            tmp = e + ": " + str(c)
            if len(out+tmp) > CHAR_LIMIT:
                await self.client.send_message(self.message.channel, out)
                out = tmp
            else:
                out += tmp

        return out

class Chungus(Emoji):
    desc = ":chungus:"

    async def eval(self, emoji=None):
        if emoji == None:
            emoji_list = self.server.emojis
            if not emoji_list:
                raise CommandFailure("Emoji list is empty!")
            emoji = str(random.choice(emoji_list))

        return f"<:cw:590153701252005907><:c1_0:590153698324381698><:c2_0:590153703609204760><:cw:590153701252005907>\n<:cw:590153701252005907><:c1_1:590153699372826634>{emoji}<:c3_1:590153701281497109>\n<:c0_2:590153690493747220><:c1_2:590153704129298442><:c2_2:590153702443319307><:c3_2:590153704121040898>\n<:c0_3:590153695363203072><:c1_3:590153703454015493><:c2_3:590153701969231873><:c3_3:590153703697416192>"


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
