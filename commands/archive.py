from commands.base import Command
from helpers import *
from discord import Embed, NotFound, Forbidden, HTTPException
from configstartup import config

import asyncio
import datetime

HISTORY = 10
HISTORY_LIMIT = 500
SCROLL_UTF = "\U0001F4DC"
HOTLINK_PREFIX = "https://discordapp.com/channels/"
ARCHIVE_CHANNEL = config['CHANNELS'].get('Archive')
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
IMAGE_EXTENSIONS = ['png', 'jpg', 'jpeg', 'gif']


class Archive(Command):
    desc = "Archive command to store the best of PCSoc. Mods only."
    roles_required = ['mod', 'exec']

    async def eval(self, index):
        try:
            if int(index) >= HISTORY:
                message = await self.client.get_message(self.message.channel, index)
                entry = Entry(index, message)
            else:
                archive = await create_archive(self.client.logs_from(self.message.channel,
                                                                     limit=HISTORY_LIMIT))
                entry = archive[int(index)]
            channel = self.client.get_channel(ARCHIVE_CHANNEL)
            header = entry.hotlink
            footer = "Archived message from #{}".format(self.message.channel)
            await self.client.send_message(channel, header, embed=entry.as_embed(footer))
            return "Archived message {} in <#{}>".format(index, ARCHIVE_CHANNEL)

        except (IndexError, ValueError, NotFound, Forbidden, HTTPException):
            return "Could not archive message %s!" % bold(index)


class List(Archive):
    desc = "Lists recent messages available for archiving. Mods only."

    async def eval(self):
        archive = await create_archive(self.client.logs_from(
            self.message.channel, limit=HISTORY_LIMIT))
        out = [entry.as_text() for entry in archive]
        if out:
            return SCROLL_UTF + "Last %d Archiveable Messages:\n" % len(out) + "\n".join(out)
        return "No messages available for archival. React with a " + SCROLL_UTF \
            + " to a message in the last " + str(HISTORY_LIMIT) + " messages of"\
            + " this channel to mark it for archival."


class Ls(Archive):
    desc = "See " + bold(code("!archive") + " " + code("list")) + "."

    async def eval(self):
        return await List.eval(self)


class Entry():
    def __init__(self, index, message):
        self.index = index
        reactions = [x.emoji for x in message.reactions]
        try:
            self.reactions = message.reactions[reactions.index(
                SCROLL_UTF)].count
        except ValueError:
            self.reactions = 0
        self.author = message.author
        self.content = message.clean_content
        self.attachments = message.attachments
        self.timestamp = message.timestamp
        self.hotlink = "{}{}/{}/{}".format(HOTLINK_PREFIX,
                                           message.server.id, message.channel.id, message.id)

    def as_text(self):
        # Show index
        text = str(self.index) + ". "

        # Show reaction count
        text += "[%s%d]  " % (SCROLL_UTF, self.reactions)

        # Show author
        text += nick(self.author)

        # Bold entire first line
        text = bold(text)
        text += "\n"

        # Show the content if it exists
        if self.content:
            text += bold("Content: ")
            text += code(self.content) + "\n"

        # Show attachments if they exist
        if self.attachments:
            text += bold("Attachment(s): ")
            attachments = [noembed(x["url"]) for x in self.attachments]
            text += "\n".join(attachments) + "\n"

        # Show timestamp of message
        text += bold("Timestamp: ")
        text += code(self.timestamp.strftime("%d/%m/%y %H:%M:%S")) + "\n"

        # Add the hotlink without embed
        text += bold("Hotlink: ")
        text += noembed(self.hotlink)

        return text

    def as_embed(self, title):
        embed = Embed(description=self.content,
                      colour=self.author.colour,
                      timestamp=self.timestamp)
        embed.set_author(name=nick(self.author),
                         icon_url=AVATAR_FORMAT.format(self.author))
        embed.set_footer(text=title)
        attached_image = False
        for a in self.attachments:
            if is_image(a["url"]):
                embed.set_image(url=a["url"])
                attached_image = True
            else:
                embed.add_field(name="Attachment",
                                value=a["url"], inline=False)
        if not attached_image:
            if is_image(self.content):
                embed.set_image(url=self.content)

        return embed


async def create_archive(logs):
    i = 0
    archive = []

    # Iterate through message history
    async for message in logs:
        # Break when we've found HISTORY messages
        if i >= HISTORY:
            break

        reactions = [x.emoji for x in message.reactions]
        # Find those messages that have scroll emoji reactions
        if SCROLL_UTF in reactions:

            # Append entry to archive
            archive.append(Entry(i, message))

            # Increment history counter
            i += 1

    return archive


def is_image(url):
    return url.split('.')[-1].lower() in IMAGE_EXTENSIONS
