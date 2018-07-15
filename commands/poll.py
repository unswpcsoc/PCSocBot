from commands.base import Command
from helpers import *

from discord import Embed, Emoji, NotFound, Forbidden, HTTPException

import asyncio

#DEFAULT_TITLE = "Poll: \ufe0f"
ZERO_WIDTH_SPACE = "\u200b"
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
REG_A_HEX = 'f09f87a6'
REG_A_INT = int(REG_A_HEX, base=16)

letters = []

for i in range(26):
    # fromhex() complains if there's a '0x' in the str, which hex() returns
    letters.append(bytes.fromhex(hex(REG_A_INT+i)[2:]).decode("utf8"))

class Poll(Command):
    desc = "Creates a poll with title, timeout(m), and space-separated entries"
    desc += ". Limited to 26 entries. Vote with reactions"

    async def eval(self, *entries):
        if not entries:
            raise CommandFailure("Please add some entries!")
        elif len(entries) > 26:
            raise CommandFailure("Too many entries!")

        try:
            # Construct the title and footer
            embed = Embed(colour=self.message.author.colour,
                          #title=DEFAULT_TITLE,
                          timestamp=self.message.timestamp)

            embed.set_author(name=nick(self.message.author),
                            icon_url=AVATAR_FORMAT.format(self.message.author))

            # TODO calculate ending time
            #embed.set_footer(text="Voting ends at %s" % end_time)

            # Construct entries
            i = 0
            for entry in entries:
                embed.add_field(name=letters[i], value=entry, inline=False)
                i += 1

            # Send embed to the current channel
            msg = await self.client.send_message(self.message.channel, embed=embed)

            # React to the sent message
            for i in range(len(entries)):
                await self.client.add_reaction(msg, letters[i])

        except (NotFound, Forbidden, HTTPException):
            CommandFailure("Could not construct poll!")
