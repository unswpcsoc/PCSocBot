#!/usr/bin/python3

import json
import os
import sys

import discord

import commands
#from commands.highnoon import high_noon, HIGH_NOON_CHANNEL
from commands.leaderboard import leaderboard, LEADERBOARD_CHANNEL
from commands.twitch import twitch, TWITCH_CHANNEL

client = discord.Client()
high_noon_channel = None

err = """OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo!
The code monkeys at our headquarters are working VEWY HAWD to fix this!"""

@client.event
async def on_ready():
    # Set game by CLA or default
    presence = sys.argv[1] if len(sys.argv) == 2 else "Despacito 2"

    print('------')
    print('Logged in as')
    print(client.user.name)
    print(client.user.id)
    print("Playing " + presence)
    print('------')

    await client.change_presence(game=discord.Game(name=presence))
    for channel in client.get_all_channels():

        #if channel.name == HIGH_NOON_CHANNEL:
            #await high_noon(client, channel)

        # if channel.name == LEADERBOARD_CHANNEL:
        #     await leaderboard(client, channel)

        if channel.name == TWITCH_CHANNEL:
            await twitch(client, channel)

@client.event
async def on_message(message):
    try:
        if message.content.startswith(commands.PREFIX):
            args = message.content[1:].split()
            if args:
                cls, args = commands.Help.find_command(args)
                output = await cls(client, message).init(*args)
                if isinstance(output, discord.Embed):
                    await client.send_message(message.channel, embed=output)
                elif output is not None:
                    await client.send_message(message.channel, output)
    except discord.errors.HTTPException as e:
        await client.send_message(message.channel, err)

client.run(os.environ['TOKEN'])
