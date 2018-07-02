from commands.base import Command
from helpers import *

import asyncio

HISTORY = 10
HISTORY_LIMIT = 500
SCROLL_UTF = "\U0001F4DC"
HOTLINK_PREFIX = "https://discordapp.com/channels/"
ARCHIVE_CHANNEL = "462063414408249376"

class Archive(Command):
    desc = "Archive command to store the best of PCSoc. Mods only."
    async def eval(self, index):
        out = await create_archive(self.client.logs_from(self.message.channel))
        channel = self.client.get_channel(ARCHIVE_CHANNEL)
        await self.client.send_message(channel, out[int(index)])
        return "Archived message {} in <#{}>".format(index, ARCHIVE_CHANNEL)

class List(Archive):
    desc = "Lists recent messages available for archiving. Mods only."
    async def eval(self):
        out = await create_archive(self.client.logs_from(self.message.channel))
        return SCROLL_UTF + "Last 10 Archiveable Messages:\n" + "\n".join(out)

async def create_archive(logs):
    i = 0
    out = []

    # Iterate through message history
    async for message in logs:
        # Break when we've found HISTORY messages
        if i >= HISTORY:
            break

        reactions = [x.emoji for x in message.reactions]
        # Find those messages that have scroll emoji reactions
        if SCROLL_UTF in reactions:

            # Enumerate archivable posts
            inner = str(i) + ". "

            # Show reaction count
            sformat = (SCROLL_UTF,
                    message.reactions[reactions.index(SCROLL_UTF)].count)
            inner += "[%s%d]  " % sformat

            # Show author
            inner += str(message.author)

            # Bold entire first line
            inner = bold(inner)
            inner += "\n"

            # Show the content if it exists
            if message.clean_content:
                inner += bold("Content: ")
                inner += code(message.clean_content) + "\n"

            # Show attachments if they exist
            if message.attachments:
                inner += bold("Attachment(s): ")
                attachments = message.attachments
                attachments = [ noembed(x["url"]) for x in attachments ]
                inner += "\n".join(attachments) + "\n"

            # Add the hotlink without embed
            inner += bold("Hotlink: ") + noembed(HOTLINK_PREFIX 
                    + message.server.id + "/" 
                    + message.channel.id + "/" 
                    + message.id)+ "\n"

            # Append inner to out
            out.append(inner)

            # Increment history counter
            i += 1
    
    return out