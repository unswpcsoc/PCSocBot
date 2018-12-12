from commands.base import Command
from helpers import *
from discord import Embed
from utils.username_generator import *
import requests

REPORT_CHANNEL = 'report'
IMAGE_EXTENSIONS = ['png', 'jpg', 'jpeg', 'gif']

non_member = "You are not a member of the PCSoc Discord 🙁"
help_message = 'Any private message sent to the bot will be forwarded anonymously to a mod only text channel on the UNSW PCSoc Discord Server'
success_message = "Thank you for your message, it has been forwarded anonymously to the PCSoc Moderation Team 🙂"
report_message = '❗ @ everyone new mod report ❗'

report_authors = dict()

class Report(Command):
    desc = "Anonymous mod reporting for the PCSoc Discord Server. PM the bot to make a report."


class Reply(Report):
    desc = "Replies to a mod report."
    roles_required = ["mod"]
    async def eval(self, nickname, *message):
        if not nickname.isalpha():
            raise CommandFailure("Nickname is invalid!")
        
        report_author = report_authors.get(nickname.lower())
        if report_author is None:
            raise CommandFailure("Nickname does not exist!")

        message = ' '.join(message)
        if not message:
            raise CommandFailure("Message cannot be blank!")

        channel = None
        for member in self.server.members:
            if report_author == member.id:
                channel = member
                break
        if channel is None:
            CommandFailure("Member is no longer in the server!")

        reply_message = "A mod has replied: " + message

        await self.client.send_message(channel, reply_message)

        return "Report reply sent!"


async def report(client, channel, message):
    # checks if the message is a PM and not from a bot account
    if message.server is None and not message.author.bot:
        # checks if user is in server
        if message.author not in channel.server.members:
            await client.send_message(message.author, non_member)
            return True

        #returns help message if requested
        if message.content.startswith('!help'):
            await client.send_message(message.author, help_message)
            return True

        #generate username and store the reverse mapping locally
        name = get_uname(message.author.id)
        global report_authors
        report_authors[name.lower()] = message.author.id

        #construct and send embed
        colour = get_ucolour(message.author.id)
        embed = Embed(description=message.content, colour=colour, timestamp=message.timestamp)
        embed.set_author(name=name)

        attached_image = False
        for a in message.attachments:
            if is_image(a["url"]):
                embed.set_image(url=a["url"])
                attached_image = True
            else:
                embed.add_field(name="Attachment", value=a["url"], inline=False)
        if not attached_image:
            if is_image(message.content):
                embed.set_image(url=message.content)

        await client.send_message(channel, report_message, embed=embed)
        await client.send_message(message.author, success_message)

        return True

    return False


def is_image(url):
    return url.split('.')[-1].lower() in IMAGE_EXTENSIONS and url_exists(url)


def url_exists(path):
    try:
        r = requests.head(path)
        return r.status_code == requests.codes.ok
    except requests.RequestException:
        return False
