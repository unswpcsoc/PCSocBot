from commands.base import Command
from helpers import *


class Weeb(Command):
    desc = "Gives/removes the Weeb role."
    
    async def eval(self):
        return await assign_role(self.client, self.message, "Weeb")


class Wiki(Command):
    desc = "Gives/removes the Wiki Gremlin role."

    async def eval(self):
        return await assign_role(self.client, self.message, "Wiki Gremlin")


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
        raise CommandFailure(f"{role_name} role does not exist!")

    if role in message.author.roles:
        await client.remove_roles(message.author, role)
        return f"{message.author.mention} is no longer a {role_name}."

    await client.add_roles(message.author, role)
    return f"{message.author.mention} is now a {role_name}."


def find_role(roles, role_name):
    for role in roles:
        if role_name == role.name:
            return role
    return None
