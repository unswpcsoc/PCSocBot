from commands.base import Command
from helpers import *
from discord.utils import get
from configstartup import config


class Roles(Command):
    role_name = ""
    desc = "Gives/removes a role."
    mesg = ""
    removable = True

    async def eval(self):
        if not self.role_name:
            return self.help
            
        role = self.find_role(self.message.server.roles, self.role_name)
        if role is None:
            raise CommandFailure("%s role does not exist!" % self.role_name)

        if role in self.message.author.roles:
            if not removable:
                raise CommandFailure("Role cannot be removed!")

            await self.client.remove_roles(self.message.author, role)
            return self.message.author.mention + " is no longer a " + self.role_name

        await self.client.add_roles(self.message.author, role)
        if self.mesg:
            await self.client.send_message(self.message.author, mesg)
        return self.message.author.mention + " is now a " + self.role_name


    def find_role(self, roles, role_name):
        for role in roles:
            if role_name == role.name:
                return role
        return None


class Accept(Roles):
    role_name = "Literate"
    desc = "Gives the Literate role. Cannot be removed."
    mesg = ""
    removable = False
    alias_names = ['literate']


class Weeb(Roles):
    role_name = "Weeb"
    desc = "Gives/removes the Weeb role."
    mesg = ""


class Wiki(Roles):
    role_name = "Wiki Gremlin"
    desc = "Gives/removes the Wiki Gremlin role."
    mesg = ""


class Meta(Roles):
    role_name = "Meta"
    desc = "Gives/removes the Meta role."
    mesg = ""


class Bookworm(Roles):
    role_name = "Bookworm"
    desc = "Gives/removes the Bookworm role."
    mesg = ""
