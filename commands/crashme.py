from commands.base import Command


class Crashme(Command):
    desc = "Logs bot out of discord. Mod only."
    roles_required = ["mod", "exec"]

    async def eval(self):
        await self.client.logout()
