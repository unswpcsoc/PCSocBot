from commands.base import Command
from helpers import *
import urllib.request
import json
import re
import asyncio
import os
import time
import datetime
import configparser
from utils.embed_table import EmbedTable
from discord import Embed
from configstartup import config

TWITCH_CHANNEL = config['CHANNELS'].get('Stream')
TWITCH_FILE = config['FILES'].get('Twitch')
TWITCH_COLOR = int('6441a4', 16)

HEADERS = {'Accept': 'application/vnd.twitchtv.v5+json',
           'Client-ID': config['KEYS'].get('TwitchClientID')}
SLEEP_INTERVAL = 300
REQUEST_PREFIX = 'https://api.twitch.tv/kraken/'


class Twitch(Command):
    desc = "Twitch go live alerts for #stream"


class Add(Twitch):
    desc = "Adds user by Twitch channel name. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, username):
        # check twitch channel name is valid
        pattern = re.compile("^[a-zA-Z0-9_]{4,25}$")
        if pattern.match(username) is None:
            raise CommandFailure(
                code(username) + ' is not a valid Twitch username!')

        # check channel exists
        req = urllib.request.Request(REQUEST_PREFIX + 'users?login=' + username,
                                     data=None, headers=HEADERS)
        res = urllib.request.urlopen(req)
        data = json.loads(res.read().decode('utf-8'))
        if data['_total'] == 0:
            raise CommandFailure(code(username) + ' channel does not exist!')

        key = username.lower()
        name = data['users'][0]['display_name']
        id = data['users'][0]['_id']

        # Open the JSON file or create a new dict to load
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
        except FileNotFoundError:
            channels = {}
            channels['channels'] = {}

        # Add the format string to the key
        if key in channels['channels']:
            raise CommandFailure(
                code(name) + ' is already on the list of broadcasters!')
        channel = {'id': id, 'name': name}
        channels['channels'][key] = channel

        # Write the formats to the JSON file
        with open(TWITCH_FILE, 'w') as new:
            json.dump(channels, new)

        return code(name) + ' has been added to the list of broadcasters!'


class Remove(Twitch):
    desc = "Removes Twitch channel from list of broadcasters. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, username):
        # check twitch channel name is valid
        pattern = re.compile("^[a-zA-Z0-9_]{4,25}$")
        if pattern.match(username) is None:
            raise CommandFailure(
                code(username) + ' is not a valid Twitch username!')

        # Open the JSON file or create a new dict to load
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
            channels['channels'].pop(username.lower())

        except (FileNotFoundError, KeyError, ValueError):
            raise CommandFailure("Broadcaster %s not found!" % code(username))

        # Write the formats to the JSON file
        with open(TWITCH_FILE, 'w') as new:
            json.dump(channels, new)

        return code(username) + ' was removed from the list of broadcasters!'


class Rm(Twitch):
    desc = "See " + bold(code("!twitch") + " " + code("remove")) + "."
    def eval(self, username): return Remove.eval(self, username)


class List(Twitch):
    desc = "Lists the stored broadcaster channel names."

    def eval(self):
        # Open the JSON file, skip if it does not exist
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
        except FileNotFoundError:
            raise CommandFailure("Broadcaster list is empty!")

        names = [value['name'] for key, value in channels['channels'].items()]
        names = sorted(names)

        return EmbedTable(fields=['Broadcasters'],
                          table=[(name,) for name in names],
                          colour=TWITCH_COLOR)


class Ls(Twitch):
    desc = "See " + bold(code("!twitch") + " " + code("list")) + "."
    def eval(self): return List.eval(self)


class Setm(Twitch):
    desc = "Sets the custom go live message for a specified twitch channel. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, username, *message_string):
        # Get the message string
        message_string = " ".join(message_string)
        message_string = message_string.replace('\\n', '\n')

        # Open the JSON file and attempt to write the message
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
            channels['channels'][username.lower()]['message'] = message_string
            name = channels['channels'][username.lower()]['name']

        except (FileNotFoundError, KeyError, ValueError):
            raise CommandFailure("Broadcaster %s not found!" % code(username))

        # Write the formats to the JSON file
        with open(TWITCH_FILE, 'w') as new:
            json.dump(channels, new)

        return "Message for %s set to %s!" % (code(name), code(message_string))


class Removem(Twitch):
    desc = "Removes the custom go live message for a specified channel if it exists. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, username):
        # Open the JSON file and attempt to get the message
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
            channel = channels['channels'][username.lower()]
            name = channels['channels'][username.lower()]['name']

        except (FileNotFoundError, KeyError, ValueError):
            raise CommandFailure("Broadcaster %s not found!" % code(username))

        message = channel.pop('message', None)
        if message is None:
            raise CommandFailure("No custom message for %s set!" % code(name))

        # Write the formats to the JSON file
        with open(TWITCH_FILE, 'w') as new:
            json.dump(channels, new)

        return "Custom message for %s removed!" % code(name)


class Getm(Twitch):
    desc = "Gets the custom go live message for a specified channel if it exists. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, username):
        # Open the JSON file and attempt to get the message
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
            channel = channels['channels'][username.lower()]
            name = channels['channels'][username.lower()]['name']

        except (FileNotFoundError, KeyError, ValueError):
            raise CommandFailure("Broadcaster %s not found!" % code(username))

        message = channel.get('message', None)
        if message is None:
            raise CommandFailure("No custom message for %s set!" % code(name))

        return "Message for %s is %s!" % (code(name), code(message))


# Twitch Alerts Event Loop

async def twitch(client, channel):
    messages = dict()
    # Event Loop
    while True:
        # Sleep at start
        await asyncio.sleep(SLEEP_INTERVAL)

        # Open the JSON file, skip if it does not exist
        try:
            with open(TWITCH_FILE, 'r') as old:
                channels = json.load(old)
        except FileNotFoundError:
            continue

        for key, value in channels['channels'].items():
            name = value['name']
            id = value['id']
            message = value.get('message', None)

            # Check if channel is live
            try:
                req = urllib.request.Request(REQUEST_PREFIX + 'streams/' + id,
                                             data=None, headers=HEADERS)
                res = urllib.request.urlopen(req)
                data = json.loads(res.read().decode('utf-8'))
                stream = data['stream']
            except urllib.error.URLError as e:
                print(timestamp() + ' TWITCH: "' + str(e) + '" for ' + name)
                continue

            # skip if channel is not live
            if not stream:
                messages.pop(key, None)  # remove stored message
                continue

            # set message
            if message is None:
                body = 'Hey guys, %s is now live on %s ! Go check it out!' % \
                    (code(name), stream['channel']['url'])
            else:
                body = message
            description = '[%s](%s)' % \
                (stream['channel']['status'], stream['channel']['url'])
            ts = datetime.datetime.utcnow()
            footer = 'Last updated'

            # set embed contents
            icon = stream['channel'].get('logo', '')
            image = stream['preview'].get('large', '')
            game = stream['channel']['game']
            game = game if len(game) > 0 else 'No Game Specified'
            viewers = stream.get('viewers', '')

            embed = Embed(description=description,
                          timestamp=ts, colour=TWITCH_COLOR)

            embed.set_author(name=name, icon_url=icon)

            embed.set_image(url=image + '?time='+str(int(time.time())))

            embed.set_thumbnail(url=icon)

            embed.set_footer(text=footer)

            embed.add_field(name='Game', value=game, inline=True)
            embed.add_field(name='Viewers', value=viewers, inline=True)

            # edit existing message if it exists, or create new message
            if key in messages:
                await client.edit_message(messages[key], embed=embed)

            else:
                message = await client.send_message(channel, body, embed=embed)

                # store message in messages
                messages[key] = message
