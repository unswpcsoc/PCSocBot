from commands.base import Command
from helpers import *

from discord import Embed, Emoji, NotFound, Forbidden, HTTPException

import asyncio, datetime

REG_A_HEX = 'f09f87a6'
REG_A_INT = int(REG_A_HEX, base=16)

ENTRY_SEPARATOR = ";;"
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
DURATION = 60

class Poll(Command):
    desc = "Creates a poll using `;;`-separated entries, sorted by length. "
    desc += "Limited to 20 entries. Vote with reactions"

    async def eval(self, *entries):
        if not entries:
            raise CommandFailure("Please add some entries!")

        # Split on separator
        entries = " ".join(entries).split(ENTRY_SEPARATOR)
        entries = [ x.strip() for x in entries ]

        if len(entries) > 20:
            raise CommandFailure("Too many entries! (Max 20)")

        try:
            # Construct the title and footer
            embed = Embed(colour=self.message.author.colour,
                          #title=DEFAULT_TITLE,
                          timestamp=self.message.timestamp)

            embed.set_author(name=nick(self.message.author),
                            icon_url=AVATAR_FORMAT.format(self.message.author))

            # Construct entries
            i = 0
            entries.sort(key=len, reverse=True)
            for entry in entries:

                # Remove empty entries
                if not entry: 
                    entries.pop(i) 
                    continue

                # Construct entry
                embed.add_field(name=letters(i), value=entry, inline=True)
                i += 1

            # TODO Make absolute time
            duration = str(datetime.timedelta(seconds=int(DURATION)))
            embed.set_footer(text="Votes counted in [%s]" % duration)

            # Send embed to the current channel
            msg = await self.client.send_message(self.message.channel, \
                                                 embed=embed)

            # React to the sent message
            for i in range(len(entries)):
                await self.client.add_reaction(msg, letters(i))
            
            # Delay vote counting
            await asyncio.sleep(DURATION)

            # Get the new cached message instead of the temporary one
            new = await self.client.get_message(self.message.channel, msg.id)

            # Convert in-place to index from regional_a
            # unicode to hex to int rep to index from regional_a
            votes = []
            for x in new.reactions:
                try:
                    votes.append((x.count, \
                        int(x.emoji.encode('utf8').hex(), \
                        base=16) - REG_A_INT))
                except AttributeError: continue

            # Get only the reactions within A-T
            votes = [ x for x in votes if 0 <= x[1] < 20 ]

            # Sort by highest vote, then length of original entry
            votes.sort(key=lambda x: (-x[0], -len(entries[x[1]])))

            # Reconstruct embed
            embed = Embed(colour=self.message.author.colour,
                          #title=DEFAULT_TITLE,
                          timestamp=self.message.timestamp)

            embed.set_author(name=nick(self.message.author),
                            icon_url=AVATAR_FORMAT.format(self.message.author))

            # Construct results
            for vote in votes:
                name = "%s %d votes" % (letters(vote[1]), vote[0])
                value = entries[vote[1]]
                embed.add_field(name=name, value=value, inline=True)

            await self.client.send_message(self.message.channel, embed=embed)

        except (NotFound, Forbidden, HTTPException):
            CommandFailure("Could not construct poll!")


def letters(index):
    """ Converts the index into the corresponding bytestring for the 
    unicode regional indicator emoji character """

    if not 0 <= index < 26:
        raise IndexError
    else:
        # Return the unicode bytes rep for the letter
        # fromhex() complains if there's a '0x' in the str, which hex() returns
        return bytes.fromhex(hex(REG_A_INT+index)[2:]).decode("utf8")
