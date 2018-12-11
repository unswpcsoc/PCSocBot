from datetime import datetime
import json
import random
import re

from discord import Embed

from commands.base import Command
from helpers import *

QUOTE_FILE = "files/quotes.json"
PENDING_FILE = "files/pending.json"
CHAR_LIMIT = 2000
LIST_LIMIT = 50
DEFAULT_COLOR = int('000000', 16)


class Quote(Command):
    desc = "Quote storage and retrieval system. Retrieve by index, or leave blank for random."

    async def eval(self, index=-1):
        # convert input index to int
        try:
            index = int(index)
        except ValueError:
            raise CommandFailure(
                "Must supply either an integer or subcommand!")

        # Retrieve the quotes list
        try:
            # Get the format dict, throws FileNotFoundError
            with open(QUOTE_FILE, 'r') as fmt:
                quotes = json.load(fmt)
        except FileNotFoundError:
            raise CommandFailure('Quotes list is empty!')

        # failure conditions
        if not quotes['quotes']:
            raise CommandFailure('Quotes list is empty!')

        if index > quotes['last_id']:
            raise CommandFailure('Index out of range!')

        # pick a quote
        if index < 0:
            choice = random.choice(list(quotes['quotes'].items()))
            index = choice[0]
            quote = choice[1]
        elif str(index) in quotes['quotes']:
            quote = quotes['quotes'][str(index)]
        else:
            raise CommandFailure(f'Quote with ID {index} does not exist!')

        # find user and use details, or if they have left the server, use the default
        user = self.from_id(quote['author'])
        if user is None:
            name = quote['nick']
            colour = DEFAULT_COLOR
        else:
            name = user.name
            colour = user.colour.value

        # construct and send embed
        message = ''
        title = f'Quote #{index}'
        body = quote['quote']
        footer = f'Added by {name}'
        timestamp = datetime.strptime(
            quote['timestamp'], '%Y-%m-%d %H:%M:%S.%f')

        embed = Embed(description=body, colour=colour, timestamp=timestamp)
        embed.set_author(name=title)
        embed.set_footer(text=footer)

        await self.client.send_message(self.message.channel, message, embed=embed)


class Add(Quote):
    desc = "Requests a quote to be added to the bot, pending mod approval."

    def eval(self, *quote_string):
        # Get the format string
        quote_string = " ".join(quote_string)
        quote_string = quote_string.replace('\\n', '\n')

        # Open the JSON file or create a new dict to load
        try:
            with open(PENDING_FILE, 'r') as old:
                pending = json.load(old)
        except FileNotFoundError:
            pending = []

        quote = {
            'quote': quote_string,
            'author': self.user,
            'nick': self.name,
            'timestamp': str(self.message.timestamp)
        }

        # Add the quote string to the pending list
        pending.append(quote)

        # Write the quotes list to the JSON file
        with open(PENDING_FILE, 'w') as new:
            json.dump(pending, new)

        return f'The following quote has been added to the pending list at index ' \
            f'{len(pending)-1}:\n{codeblock(quote_string)}'


class Remove(Quote):
    desc = "Removes a quote by index from the quotes list and puts it back on the pending list. Mods only."
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
            raise CommandFailure('Quotes list is empty!')

        if index < 0 or index > quotes['last_id']:
            raise CommandFailure('Index out of range!')

        try:
            quote = quotes['quotes'].pop(str(index))
        except KeyError:
            raise CommandFailure(f'Quote with ID {index} does not exist!')

        # Open the pending file or create a new dict to load
        try:
            with open(PENDING_FILE, 'r') as old:
                pending = json.load(old)
        except FileNotFoundError:
            pending = []

        # Add the quote string to the pending list
        pending.append(quote)

        # Write the quotes list to the JSON file
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        # Write the pending list to the JSON file
        with open(PENDING_FILE, 'w') as new:
            json.dump(pending, new)

        return f'Quote with ID {index} moved to pending list!'


class Approve(Quote):
    desc = "Approves a quote from the pending list and puts it in the quotes list. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, index):
        # convert input index to int
        try:
            index = int(index)
        except ValueError:
            raise CommandFailure(
                "Must supply either an integer or subcommand!")

        # Retrieve the pending list
        try:
            # Get the format dict, throws FileNotFoundError
            with open(PENDING_FILE, 'r') as fmt:
                pending = json.load(fmt)
        except FileNotFoundError:
            raise CommandFailure('Pending list is empty!')

        if not pending:
            raise CommandFailure('Pending list is empty!')

        if index < 0 or index >= len(pending):
            raise CommandFailure('Index out of range!')

        # copy and remove from the pending list
        quote = pending.pop(index)

        # retrieve the quotes list
        try:
            with open(QUOTE_FILE, 'r') as old:
                quotes = json.load(old)
        except FileNotFoundError:
            quotes = {}

        # get ID of last entry
        if not quotes:  # quotes dict is empty
            next_id = 0
            quotes['quotes'] = {}
        else:
            next_id = quotes['last_id']+1

        quotes['quotes'][next_id] = quote
        quotes['last_id'] = next_id

        # Write to the quote JSON
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        # Write to the pending JSON
        with open(PENDING_FILE, 'w') as new:
            json.dump(pending, new)

        user = self.from_id(quote['author'])
        if user is None:
            name = quote['nick']
        else:
            name = user.mention

        return f"The following quote by {name} has been added to the quotes list with ID " \
            f"{next_id}:\n{codeblock(quote['quote'])}"


class Reject(Quote):
    desc = "Rejects a quote by index from the pending list. Mods only."
    roles_required = ['mod', 'exec']

    def eval(self, index):
        # convert input index to int
        try:
            index = int(index)
        except ValueError:
            raise CommandFailure("Must supply an integer!")

        # Open the JSON file or create a new dict to load
        try:
            with open(PENDING_FILE, 'r') as old:
                pending = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('Pending list is empty!')

        if index < 0 or index >= len(pending):
            raise CommandFailure('Index out of range!')

        quote = pending.pop(index)

        # Write the formats to the JSON file
        with open(PENDING_FILE, 'w') as new:
            json.dump(pending, new)

        return f'Pending quote with index {index} removed!'


class List(Quote):
    desc = "Lists the first 50 characters of all quotes."

    async def eval(self):
        # Open the JSON file or create a new dict to load
        try:
            with open(QUOTE_FILE, 'r') as old:
                quotes = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('Quotes list is empty!')

        if not quotes['quotes']:
            raise CommandFailure('Quotes list is empty!')

        quotes = list(quotes['quotes'].items())
        quotes.sort(key=lambda tup: int(tup[0]))

        # print list of quotes
        out = '**List of Quotes:**\n'
        for quote in quotes:
            i = quote[0]
            q = quote[1]['quote'][:LIST_LIMIT]

            tmp = f'**#{i}:** {q}'
            tmp += '[...]\n' if len(q) >= LIST_LIMIT else '\n'
            if len(out+tmp) > CHAR_LIMIT:
                await self.client.send_message(self.message.channel, out)
                out = tmp
            else:
                out += tmp

        return out


class Pending(Quote):
    desc = "Displays a list of pending quotes. Mods only."
    roles_required = ['mod', 'exec']

    async def eval(self):
        # Open the JSON file or create a new dict to load
        try:
            with open(PENDING_FILE, 'r') as old:
                pending = json.load(old)
        except FileNotFoundError:
            raise CommandFailure('Pending list is empty!')

        if not pending:
            raise CommandFailure('Pending list is empty!')

        # print list of quotes
        out = '**List of *Pending* Quotes:**\n'
        for i in range(len(pending)):
            user = self.from_id(pending[i]['author'])
            a = pending[i]['nick'] if user is None else user.name
            q = pending[i]['quote']
            tmp = f'**#{i} by {a}:** {q}\n'
            if len(out+tmp) > CHAR_LIMIT:
                await self.client.send_message(self.message.channel, out)
                out = tmp
            else:
                out += tmp

        return out


class Changeid(Quote):
    desc = "Changes the ID of a quote on the quotes list. Mods only."
    roles_required = ['mod', 'exec']

    async def eval(self, oldid, newid):
        # convert input index to int
        try:
            oldid = int(oldid)
            newid = int(newid)
        except ValueError:
            raise CommandFailure(
                "Must supply either an integer or subcommand!")

        # Retrieve the quotes list
        try:
            # Get the format dict, throws FileNotFoundError
            with open(QUOTE_FILE, 'r') as fmt:
                quotes = json.load(fmt)
        except FileNotFoundError:
            raise CommandFailure('Quotes list is empty!')

        # failure conditions
        if len(quotes['quotes']) == 0:
            raise CommandFailure('Quotes list is empty!')

        if oldid > quotes['last_id'] or newid > quotes['last_id']:
            raise CommandFailure('Index out of range!')

        # check that quote with ID newid does not exist
        if str(newid) in quotes['quotes']:
            raise CommandFailure(f'Quote with ID {newid} already exists!')

        # get quote with ID oldid
        try:
            quote = quotes['quotes'].pop(str(oldid))
        except KeyError:
            raise CommandFailure(f'Quote with ID {oldid} does not exist!')

        # add quote back with ID newid
        quotes['quotes'][str(newid)] = quote

        # Write the quotes list to the JSON file
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        return f'ID for quote {oldid} changed to {newid}!'


class Ls(Quote):
    desc = f"See {bold(code('!quote'))} {bold(code('list'))}."

    async def eval(self):
        return await List.eval(self)
