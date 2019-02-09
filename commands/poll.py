from commands.base import Command
from helpers import *

from discord import Embed, Emoji, NotFound, Forbidden, HTTPException

import asyncio
import datetime

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

        entries = " ".join(entries).split(ENTRY_SEPARATOR)
        entries = [x.strip() for x in entries]

        # Check for duration argument
        if entries[0].lower == "duration":
            try:
                float(entries[1])
                return Duration.eval(self, entries[1])
            except ValueError:
                pass

        if len(entries) > 20:
            raise CommandFailure("Too many entries! (Max 20)")

        # Remove empty entries
        i = 0
        for entry in entries:
            if not entry:
                entries.pop(i)
            i += 1

        try:
            # Construct embed
            # Pad final time by time it takes to react
            # which is, at worst, one second per react
            author = self.message.author
            embed = Embed(title="Poll:",
                          colour=author.colour,
                          timestamp=self.message.timestamp +
                          datetime.timedelta(seconds=DURATION+len(entries)))

            # Add author
            embed.set_author(name=nick(author),
                             icon_url=AVATAR_FORMAT.format(author))

            # Add entries
            i = 0
            entries.sort(key=len, reverse=True)
            for entry in entries:
                # Construct entry
                embed.add_field(name=letters(i), value=entry, inline=True)
                i += 1

            # Add duration until vote counting
            duration = str(datetime.timedelta(seconds=int(DURATION)))
            embed.set_footer(text="Votes counted in [%s]" % duration)

            # Send embed to the current channel
            msg = await self.client.send_message(self.message.channel,
                                                 embed=embed)

            # React to the sent message
            for i in range(len(entries)):
                await self.client.add_reaction(msg, letters(i))

            # Delay vote counting
            await asyncio.sleep(DURATION)

            # Get the new cached message instead of the temporary one
            new = await self.client.get_message(self.message.channel, msg.id)

            # Convert in-place to index from regional_a
            # unicode char -> hex rep -> int rep -> index from regional_a
            votes = []
            for x in new.reactions:
                try:
                    votes.append((x.count-1,
                                  int(x.emoji.encode('utf8').hex(),
                                      base=16) - REG_A_INT))
                except AttributeError:
                    continue

            # Get only the reactions within A-T
            votes = [x for x in votes if 0 <= x[1] < 20]

            # Sort by highest vote, then length of original entry
            votes.sort(key=lambda x: (-x[0], -len(entries[x[1]])))

            # Reconstruct embed
            embed = Embed(colour=author.colour,
                          title="Results:",
                          timestamp=self.message.timestamp)

            embed.set_author(name=nick(author),
                             icon_url=AVATAR_FORMAT.format(author))

            # Construct results
            for vote in votes:
                name = "%s %d votes" % (letters(vote[1]), vote[0])
                value = entries[vote[1]]
                embed.add_field(name=name, value=value, inline=True)

            await self.client.send_message(self.message.channel, embed=embed)

        except (NotFound, Forbidden, HTTPException):
            raise CommandFailure("Could not construct poll!")


class Duration(Poll):
    roles_required = ["mod", "exec"]
    desc = "Changes the duration (min) of the poll before votes are" +\
           " counted. Mods only."

    def eval(self, duration):
        global DURATION

        try:
            int(duration)
        except ValueError:
            raise CommandFailure("Please enter a valid number of minutes!")

        duration = int(duration)

        if duration < 0:
            raise CommandFailure("Please enter a valid number of minutes!")

        # Get minutes
        DURATION = duration * 60

        # Return confirmation
        dur = str(datetime.timedelta(seconds=int(DURATION)))
        return "Poll duration changed to %s" % dur


def letters(index):
    """ Converts the index into the corresponding bytestring for the 
    unicode regional indicator emoji character """

    if not 0 <= index < 26:
        raise IndexError
    else:
        # Return the unicode bytes rep for the letter
        # fromhex() complains if there's a '0x' in the str, which hex() returns
        return bytes.fromhex(hex(REG_A_INT+index)[2:]).decode("utf8")
