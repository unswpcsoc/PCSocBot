from commands.base import Command
from helpers import *
from discord.utils import get
from configstartup import config


class Weeb(Command):
    desc = "Gives/removes the Weeb role."

    async def eval(self):
        return await assign_role(self.client, self.message, "Weeb")


class Meta(Command):
    desc = "Gives/removes the Meta role."

    async def eval(self):
        return await assign_role(self.client, self.message, "Meta")


class Bookworm(Command):
    desc = "Gives/removes the Bookworm role."

    async def eval(self):
        return await assign_role(self.client, self.message, "Bookworm")


async def assign_role(client, message, role_name):
    role = find_role(message.server.roles, role_name)
    if role is None:
        raise CommandFailure("%s role does not exist!" % role_name)

    if role in message.author.roles:
        await client.remove_roles(message.author, role)
        return message.author.mention + " is no longer a " + role_name

    await client.add_roles(message.author, role)
    return message.author.mention + " is now a " + role_name


def find_role(roles, role_name):
    return get(roles, id=config['ROLES'].get(role_name))
