from commands.base import Command
from helpers import *

import asyncio

HISTORY = 10
HISTORY_LIMIT = 500

class Archive(Command):

    async def eval(self):
        i = 0
        out = []
        async for message in self.client.logs_from(self.message.channel):
            if i >= HISTORY:
                break

            if message.channel == self.message.channel:
                out.append(bold(nick(message.author)) + " : " + message.clean_content)
                i += 1

        return "\n".join(out)

