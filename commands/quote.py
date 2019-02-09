from commands.base import Command
from helpers import *
from discord import Embed
from datetime import datetime
from configstartup import config

import random
import json
import re

QUOTE_FILE = config['FILES'].get('QuoteData')
PENDING_FILE = config['FILES'].get('QuotePending')
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
        if len(quotes['quotes']) == 0:
            raise CommandFailure('Quotes list is empty!')

        if index > quotes['last_id']:
            raise CommandFailure('Index out of range!')

        # pick a quote
        if index < 0:
            choice = random.choice(list(quotes['quotes'].items()))
            index = choice[0]
            quote = choice[1]
        else:
            try:
                quote = quotes['quotes'][str(index)]
            except KeyError:
                raise CommandFailure(
                    'Quote with ID %s does not exist!' % index)

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
        title = 'Quote #%s' % index
        body = quote['quote']
        footer = 'Added by %s' % name
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

        return 'The following quote has been added to the pending list at index %s:\n%s'\
            % (len(pending)-1, codeblock(quote_string))


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
            raise CommandFailure('Quote with ID %s does not exist!' % index)

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

        return 'Quote with ID %s moved to pending list!' % index


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

        return "The following quote by %s has been added to the quotes list with ID %s:\n%s"\
            % (name, next_id, codeblock(quote['quote']))


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

        return 'Pending quote with index %s removed!' % index


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

            tmp = '**#%s:** %s' % (i, q)
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
            tmp = '**#%s by %s:** %s\n' % (i, a, q)
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
            raise CommandFailure('Quote with ID %s already exists!' % newid)

        # get quote with ID oldid
        try:
            quote = quotes['quotes'].pop(str(oldid))
        except KeyError:
            raise CommandFailure('Quote with ID %s does not exist!' % oldid)

        # add quote back with ID newid
        quotes['quotes'][str(newid)] = quote

        # Write the quotes list to the JSON file
        with open(QUOTE_FILE, 'w') as new:
            json.dump(quotes, new)

        return 'ID for quote %s changed to %s!' % (oldid, newid)


class Ls(Quote):
    desc = "See " + bold(code("!quote") + " " + code("list")) + "."

    async def eval(self):
        return await List.eval(self)
