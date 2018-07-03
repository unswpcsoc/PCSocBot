from commands.base import Command
from helpers import *

import asyncio
import discord

class Playing(Command):
    desc = "Changes the Playing status of the bot. Mod only."
    roles_required = ["mod", "exec"]

    async def eval(self, *presence):
        presence = " ".join(presence)
        await self.client.change_presence(game=discord.Game(name=presence))
        return bold("Now Playing: " + presence)
