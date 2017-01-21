from commands.base import Command
from models import Tag


class Tags(Command):
    db_required = True
    pprint = dict(platform="platform/game")
    desc = "Player tag storage for the UNSW PCSoc discord server."


class Add(Tags):
    desc = "Adds/changes a player tag with associated platform/game to the list"
    def eval(self, platform, tag):
        Tag.create_or_update(user=self.user, platform=platform, tag=tag)
        return "%s added as %s tag for %s" % (tag, platform, self.name)


class Remove(Tags):
    desc = "Removes a user/player tag to the bot."
    def eval(self, platform):
        Tag.delete_or_err(user=self.user, platform=platform)
        return "%s tag for %s removed" % (platform, self.name)


class Get(Tags):
    desc = "Returns your own tag for a platform / game"
    def eval(self, platform):
        tag = Tag.get_or_err(user=self.user, platform=platform)
        return "The %s tag of %s is %s" % (platform, self.name, tag.tag)


class List(Tags):
    desc = "Returns a list of user tags for a specified platform"
    def eval(self, platform):
        tags = Tag.select_or_err(lambda x: x.platform == platform)
        return "tags stored for %s:\n" % platform + \
               "".join("%s [%s]" % (tag.tag, self.get_member(tag.user).name) for tag in tags)