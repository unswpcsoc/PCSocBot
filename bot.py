import json

import discord

import commands
from commands.highnoon import high_noon, HIGH_NOON_CHANNEL

client = discord.Client()
high_noon_channel = None

@client.event
async def on_ready():
    print('Logged in as')
    print(client.user.name)
    print(client.user.id)
    print('------')
    for channel in client.get_all_channels():
        if channel.name == HIGH_NOON_CHANNEL:
            await high_noon(client, channel)

@client.event
async def on_message(message):
    if message.content.startswith(commands.PREFIX):
        args = message.content[1:].split()
        cls, args = commands.Help.find_command(args)
        output = await cls(client, message).init(*args)
        if output is not None:
            await client.send_message(message.channel, output)

config = json.loads(open('config.json').read())
client.run(config['token'])

server = client.accept_invite(config['invite'])
