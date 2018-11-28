from helpers import *
from utils.username_generator import *

REPORT_CHANNEL = 'report'

response = "Thank you for your message, it has been forwarded anonymously to the PCSoc Moderation Team 🙂"

async def report(client, channel, message):
    # if the message is a PM and not from a bot account
    if message.server is None and not message.author.bot:
        # construct report message
        report_message = "New mod report from "
        report_message += bold(get_uname(message.author.id))
        report_message += ": " + message.content

        # send report message and boot reply
        await client.send_message(channel, report_message)
        await client.send_message(message.author, response)

        return True

    return False