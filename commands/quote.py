from commands.base import Command
from helpers import *
from discord import Embed
from datetime import datetime

import random
import json
import re

QUOTE_FILE = "files/quotes.json"
CHAR_LIMIT = 2000
LIST_LIMIT = 50


class Quote(Command):
    desc = "Quote storage and retrieval system. Retrieve by index, or leave blank for random."

    async def eval(self, index=-1):
        # convert input index to int
        try:
            index = int(index)
        except ValueError:
            raise CommandFailure("Must supply either an integer or subcommand!")

        # Retrieve the quotes list
        try:
            # Get the format dict, throws FileNotFoundError
            with open(QUOTE_FILE, 'r') as fmt:
                quotes = json.load(fmt)
        except FileNotFoundError:
            raise CommandFailure('Quotes list is empty!')

        # Choose from quotes list
        if len(quotes) == 0:
            raise CommandFailure('Quotes list is empty!')
        elif index >= len(quotes):
            raise CommandFailure('Index out of range!')
        elif index < 0:
            index = random.randint(0, len(quotes)-1)
        
        quote = quotes[index]

        message = ''
        title = 'Quote #%s' % index
        body = quote['quote']
        footer = 'Added by %s' % quote['author']
        colour = quote['colour']
        timestamp = datetime.strptime(quote['timestamp'], '%Y-%m-%d %H:%M:%S.%f')

        embed = Embed(description=body, colour=colour, timestamp=timestamp)
        embed.set_author(name=title)
        embed.set_footer(text=footer)

        await self.client.send_message(self.message.channel, message, embed=embed)


class Add(Quote):
    desc = "Adds a quote. Misuse of this command will result in a ban or server mute."

    def eval(self, *quote_string):
        # Get the format string
        quote_string = " ".join(quote_string)

        if not quote_string:
            raise CommandFailure('Please submit a none empty quote!')

        # Open the JSON file or create a new dict to load
        try:
            with open(QUOTE_FILE, 'r') as old:
                quotes = json.load(old)
        except FileNotFoundError:
            quotes = []

        quote = {
            'quote': quote_string,
            'author': self.name,
            'timestamp': str(self.message.timestamp),
            'colour': self.message.author.colour.value
        }

        # Add the quote string to the key
        quotes.append(quote)

        # Write the formats to the JSON file
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        return 'Your quote %s has been added at index %s!' % (code(quote_string), len(quotes)-1)


class Remove(Quote):
    desc = "Removes a quote by index. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, index):
        # convert input index to int
        try:
            index = int(index)
        except ValueError:
            raise CommandFailure("Must supply an integer!")

        # Open the JSON file or create a new dict to load
        try:
            with open(QUOTE_FILE, 'r') as old:
                quotes = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('List of quotes is empty!')

        if index < 0 or index >= len(quotes):
            raise CommandFailure('Invalid index!')
        
        quote = quotes.pop(index)

        # Write the formats to the JSON file
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        return 'Quote %s with index %s removed!' % (code(quote['quote']), index)


class List(Quote):
    desc = "Lists the first 50 characters of all quotes."

    async def eval(self):
        # Open the JSON file or create a new dict to load
        try:
            with open(QUOTE_FILE, 'r') as old:
                quotes = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('List of quotes is empty!')

        # print list of quotes
        out = '**List of Quotes:**\n'
        for i in range(len(quotes)):
            q = quotes[i]['quote'][:LIST_LIMIT]
            tmp = '**#%s:** %s' % (i, q)
            tmp += '[...]\n' if len(q) >= LIST_LIMIT else '\n'
            if len(out+tmp) > CHAR_LIMIT:
                await self.client.send_message(self.message.channel, out)
                out = tmp
            else:
                out += tmp
        
        return out


class Ls(Quote):
    desc = "See " + bold(code("!quote") + " " + code("list")) + "."
    async def eval(self):
        return await List.eval(self)