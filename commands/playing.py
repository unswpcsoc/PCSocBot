from commands.base import Command
from helpers import *

import asyncio
import discord

CURRENT_PRESENCE = "!helpme"


class Playing(Command):
    desc = "Changes the Playing status of the bot. Mod only."
    roles_required = ["mod", "exec"]

    async def eval(self, *presence):
        global CURRENT_PRESENCE

        CURRENT_PRESENCE = " ".join(presence)
        await self.client.change_presence(game=discord.Game(name=CURRENT_PRESENCE))
        return bold("Now Playing: " + CURRENT_PRESENCE)
