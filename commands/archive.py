import asyncio
import datetime

from discord import Embed, NotFound, Forbidden, HTTPException

from commands.base import Command
from helpers import *

HISTORY = 10
HISTORY_LIMIT = 500
SCROLL_UTF = "\U0001F4DC"
HOTLINK_PREFIX = "https://discordapp.com/channels"
ARCHIVE_CHANNEL = "474009272804442112"
AVATAR_FORMAT = "https://cdn.discordapp.com/avatars/{0.id}/{0.avatar}.png?size=128"
IMAGE_EXTENSIONS = ['png', 'jpg', 'jpeg', 'gif']


class Archive(Command):
    desc = "Archive command to store the best of PCSoc. Mods only."
    roles_required = ['mod', 'exec']

    async def eval(self, index):
        try:
            int_index = int(index)
            if int_index < 0:
                raise ValueError

            if int_index >= HISTORY:
                # Archive the Message ID given
                message = await self.client.get_message(self.message.channel, index)
                entry = Entry(index, message)
            else:
                # Archive the nth latest scroll-reacted comment
                archive = await create_archive(
                    self.client.logs_from(
                        self.message.channel, limit=HISTORY_LIMIT))
                if not archive or int_index >= len(archive):
                    raise CommandFailure(f"Can't find message {bold(index)}!")
                entry = archive[int_index]
        except ValueError:
            raise CommandFailure("Please input a non-negative number!")
        except NotFound:
            raise CommandFailure(f"Message ID {bold(index)} doesn't exist!")
        except (Forbidden, HTTPException):
            raise CommandFailure(f"Could not archive message {bold(index)}!")

        # Post archive in the archive channel
        channel = self.client.get_channel(ARCHIVE_CHANNEL)
        if channel is None:
            raise CommandFailure("Can't find archive channel!")

        footer = f"Archived message from #{self.message.channel}"
        await self.client.send_message(channel, entry.hotlink, embed=entry.as_embed(footer))
        return f"Archived message {index} in <#{ARCHIVE_CHANNEL}>"


class List(Archive):
    # BUG: If the message is greater than 2000 characters,
    # discord.py will send a 400 BAD REQUEST
    desc = "Lists recent messages available for archiving. Mods only."

    async def eval(self):
        archive = await create_archive(
            self.client.logs_from(self.message.channel,
                                  limit=HISTORY_LIMIT))
        out = [entry.as_text() for entry in archive]
        if out:
            return f"{SCROLL_UTF} Last {len(out)} Archiveable Messages:\n" + "\n".join(out)
        return f"No messages available for archival. React with a {SCROLL_UTF} " \
            f"to a message in the last {HISTORY_LIMIT} messages of " \
            f"this channel to mark it for archival."


class Ls(Archive):
    desc = f"See {bold(code('!archive'))} {bold(code('list'))}."

    async def eval(self):
        return await List.eval(self)


class Entry():
    def __init__(self, index, message):
        self.index = index
        self.reactions = 0
        # Get number of scroll reacts
        for reaction in message.reactions:
            if reaction.emoji == SCROLL_UTF:
                self.reactions = reaction.count
                break

        self.author = message.author
        self.content = message.clean_content
        self.attachments = message.attachments
        self.timestamp = message.timestamp
        self.hotlink = f"{HOTLINK_PREFIX}/{message.server.id}/{message.channel.id}/{message.id}"

    def as_text(self):
        # Show index, reaction count, author
        text = f"{self.index}. [{SCROLL_UTF}{self.reactions}]  {nick(self.author)}"

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
        for attachment in self.attachments:
            if is_image(attachment["url"]):
                embed.set_image(url=attachment["url"])
                attached_image = True
            else:
                embed.add_field(name="Attachment",
                                value=attachment["url"], inline=False)
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

        # Find those messages that have scroll emoji reactions
        if any(SCROLL_UTF == x.emoji for x in message.reactions):
            # Append entry to archive
            archive.append(Entry(i, message))

            # Increment history counter
            i += 1

    return archive


def is_image(url):
    return url.split('.')[-1].lower() in IMAGE_EXTENSIONS
