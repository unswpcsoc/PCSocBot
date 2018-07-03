from commands.base import Command
from helpers import *

import random
import json

IDENTIFIER = "{}"
FORMAT_FILE = "files/formats.json"

class Someone(Command):
    desc = "This command is used to roll for a random member of the server"

    def eval(self, people=1):
        # Create our own member list
        member_list = list(self.members)

        # Initialise vars
        roll = 0
        roll_list = []

        try:
            people = int(people)
        except ValueError:
            return "Must supply either an integer or subcommand!"

        # Assert that people is greater than 0
        people = max(people, 1)

        # Assert that the people does not exceed the members
        people = min(people, len(member_list))

        for _ in range(people):
            # Get random member from member list
            roll = random.randrange(len(member_list))

            # Append to our roll list
            rolled = member_list[roll-1]
            roll_list.append(rolled)

            # Remove the member from the member list
            member_list.remove(rolled)

        # Retrieve a random format from the json
        try:
            # Get the format dict, throws FileNotFoundError
            with open(FORMAT_FILE, 'r') as fmt:
                formats = json.load(fmt)

            # Randomly choose from the people indexing, throws KeyError
            out = random.choice(formats[str(people)])

        except (FileNotFoundError, KeyError):
            # No format for that number of people, use default
            out = "{}" + " {}"*(people-1)

        roll_list = [bold(nick(x)) for x in roll_list]

        # Craft output string according to format
        return out.format(*roll_list)

class Add(Someone):
    desc = """Adds a format template. Mods only. 
    `format_string`: String containing `people` number of {}, 
    where `{}` is replaced by a random **someone**
    """
    #roles_required = ['mod', 'exec']

    def eval(self, *format_string):
        # Get the format string
        format_string = " ".join(format_string)

        # Count the number of people to generate
        people = format_string.count(IDENTIFIER)

        # Make sure the number of people is sane
        if people < 1:
            return "Invalid format string, must have at least 1 `{}`!"

        # Open the JSON file or create a new dict to load
        try:
            with open(FORMAT_FILE, 'r') as old:
                formats = json.load(old)
        except FileNotFoundError:
            formats = {}

        # Add the format string to the key
        try:
            formats[str(people)].append(format_string)
        except KeyError:
            formats[str(people)] = [format_string]

        # Write the formats to the JSON file
        with open(FORMAT_FILE, 'w') as new:
            json.dump(formats, new)

        return "Your format %s for %s people has been added!" % (code(format_string), bold(str(people)))

class Remove(Someone):
    desc = """Removes a format template. 
    Passing a number removes all formats for those people. Mods only."""

    #roles_required = ['mod', 'exec']
    def eval(self, *format_string):

        # Get the format string
        format_string = " ".join(format_string)

        if not format_string:
            return "Must supply a format string!"

        # If a number is passed into the second argument, set remove_all flag
        try:
            people = int(format_string)
            remove_all = True
        except ValueError:
            remove_all = False
            # Get people
            people = format_string.count(IDENTIFIER)

        try:
            # Open the JSON file or create a new dict to load
            with open(FORMAT_FILE, 'r') as fmt:
                formats = json.load(fmt)

            if remove_all:
                # Remove all format strings for people
                del formats[str(people)]
                out = "Removed all formats for %s people!" % bold(str(people))

            else:
                # Remove format string if in the dict
                formats[str(people)].remove(format_string)
                out =  "The format " + code(format_string)
                out += " for " + bold(str(people))
                out += " people was removed!"


        except (FileNotFoundError, KeyError, ValueError):
            return "Format %s not found!" % code(format_string)

        # Write the formats to the JSON file
        with open(FORMAT_FILE, 'w') as new:
            json.dump(formats, new)

        return out

class List(Someone):
    desc = "Lists the formats for the given number of `people`."

    def eval(self, people=None):

        if people:
            # User has specified number of people
            try:
                int(people)
            except ValueError:
                return "`people` must be an integer value!"

        try:
            # Open the JSON file
            with open(FORMAT_FILE, 'r') as fmt:
                formats = json.load(fmt)

            if people:
                # List all the entries for people
                if not formats[people]:
                    raise NoFormatsError
                out = "Formats for " + bold(people) + " `people`:\n" 
                out += "\n".join(formats[people])
            else:
                # List all entries
                out = "All formats:\n"
                empty = True
                for k, v in sorted(formats.items()):
                    if v:
                        out += "Formats for " + bold(k) + " `people`:\n" 
                        out += "\n".join(v) + "\n"
                        empty = False 
                if empty:
                    raise NoFormatsError

        except (FileNotFoundError, KeyError, NoFormatsError):
            out = "No formats for %s people" % people if people else "No formats"

        return out

class NoFormatsError(Exception):
    pass

