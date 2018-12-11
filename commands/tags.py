from pony.orm import select, count

from commands.base import Command
from helpers import *
from models import Tag
from utils.embed_table import EmbedTable


class Tags(Command):
    db_required = True
    pprint = dict(platform="platform/game")
    desc = "Player tag storage for the UNSW PCSoc discord server."


class Add(Tags):
    desc = "Adds/changes a player tag with associated platform/game to the list."

    def eval(self, platform, tag):
        warning = ''
        if not select(count(t) for t in Tag if t.platform == platform.lower())[:][0]:
            platforms = ', '.join(sorted(select(t.platform for t in Tag)[:]))
            warning = "WARNING: creating a new platform. " \
                "Please check that the platform doesn't already exist by another name."
            warning = f"{bold(warning)}\nCurrent platforms are: {platforms}.\n"

        Tag.create_or_update(
            user=self.user, platform=platform.lower(), tag=tag)
        return f"{warning}{tag} added as {platform.title()} tag for {self.name}."


class Remove(Tags):
    desc = "Removes a user/player tag from the bot."

    def eval(self, platform):
        Tag.delete_or_err(user=self.user, platform=platform.lower())
        return f"{platform.title()} tag for {self.name} removed."


class Get(Tags):
    desc = "Returns your own tag for a specified platform/game."

    def eval(self, platform):
        tag = Tag.get_or_err(user=self.user, platform=platform.lower())
        return f"The {platform} tag of {self.name} is {tag.tag}"


class List(Tags):
    desc = "Returns a list of user tags for a specified platform."

    def eval(self, platform):
        platform = platform.lower()
        tags = Tag.select_or_err(lambda x: x.platform == platform)
        tab = []
        for tag in tags:
            try:
                tab.append((self.from_id(tag.user).name, tag.tag))
            except AttributeError:
                continue

        return EmbedTable(fields=['User', 'Tag'],
                          table=tab,
                          title=f"Showing tags for {platform.title()}",
                          colour=self.EMBED_COLOR)


class View(Tags):
    desc = "Returns a list of tags for a specified user."

    def eval(self, name):
        user = self.from_name(name)
        if user is None:
            raise CommandFailure("User not found.")
        tags = Tag.select_or_err(
            lambda x: x.user == int(user.id), "User has no tags.")
        return EmbedTable(fields=['Platform', 'Tag'],
                          table=[(x.platform.title(), x.tag) for x in tags],
                          colour=self.EMBED_COLOR, user=user,
                          title=f"Tags for {bold(user.name)}")


class Platforms(Tags):
    desc = "Returns a list of platforms."

    def eval(self):
        tags = [(platform,)
                for platform in sorted(select(x.platform for x in Tag))]
        return EmbedTable(fields=['Platform'], table=tags, colour=self.EMBED_COLOR)


class Ping(Tags):
    desc = "Pings users for a specific platform."

    def eval(self, platform):
        tags = Tag.select_or_err(lambda x: x.platform == platform)
        return ' '.join([at(tag.user) for tag in tags])


class Ask(Command):
    desc = f"See {bold(code('!tags'))} {bold(code('ping'))}."
    db_required = True
    pprint = dict(platform="platform/game")

    def eval(self, platform):
        return Ping.eval(self, platform)


class ModAdd(Tags):
    desc = "Adds/changes a specified player tag with associated platform/game to the list."
    roles_required = ["mod", "exec"]

    def eval(self, name, platform, tag):
        user = self.from_name(name)
        if user is None:
            raise CommandFailure("User not found.")

        warning = ''
        if not select(count(t) for t in Tag if t.platform == platform.lower())[:][0]:
            platforms = ', '.join(sorted(select(t.platform for t in Tag)[:]))
            warning = "WARNING: creating a new platform. " \
                "Please check that the platform doesn't already exist by another name."
            warning = f"{bold(warning)}\nCurrent platforms are: {platforms}\n"
        Tag.create_or_update(
            user=int(user.id), platform=platform.lower(), tag=tag)

        return f"{warning}{tag} added as {platform.title()} tag for {user.name}."


class ModRemove(Tags):
    desc = "Removes a specified user/player tag from the bot. Mod only."
    roles_required = ["mod", "exec"]

    def eval(self, name, platform):
        user = self.from_name(name)
        if user is None:
            raise CommandFailure("User not found.")

        Tag.delete_or_err(user=int(user.id), platform=platform.lower())
        return f"{platform.title()} tag for {user.name} removed."
