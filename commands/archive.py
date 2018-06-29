from commands.base import Command
from helpers import *

import asyncio

HISTORY = 10
HISTORY_LIMIT = 500
SCROLL_UTF = "\U0001F4DC"

class Archive(Command):

    async def eval(self):
        i = 0
        out = []
        async for message in self.client.logs_from(self.message.channel):
            if i >= HISTORY:
                break

            if SCROLL_UTF in [x.emoji for x in message.reactions]:
                i += 1
                out.append(
                        bold(nick(message.author)) + " : " 
                        + message.clean_content
                        )

        return SCROLL_UTF + "Last 10 Archiveable Messages:\n" + "\n".join(out)

