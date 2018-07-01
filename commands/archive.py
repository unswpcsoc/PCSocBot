from commands.base import Command
from helpers import *

import asyncio

HISTORY = 10
HISTORY_LIMIT = 500
SCROLL_UTF = "\U0001F4DC"
HOTLINK_PREFIX = "https://discordapp.com/channels/"

class Archive(Command):

    async def eval(self):
        i = 0
        out = []

        # Iterate through message history
        async for message in self.client.logs_from(self.message.channel):
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
                    inner += "\n".join(attachments)

                # Add the hotlink without embed
                inner += noembed(HOTLINK_PREFIX 
                        + message.server.id + "/" 
                        + message.channel.id + "/" 
                        + message.id)+ "\n"

                # Append inner to out
                out.append(inner)

                # Increment history counter
                i += 1

        return SCROLL_UTF + "Last 10 Archiveable Messages:\n" + "\n".join(out)

