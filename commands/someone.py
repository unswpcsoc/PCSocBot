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
        people = int(people)

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

        return "Your format %s for %s people has been added!" % (code(format_string), code(str(people)))

class Remove(Someone):
    desc = "Removes a format template. Mods only."

    #roles_required = ['mod', 'exec']
    def eval(self, *format_string):

        # Get the format string
        format_string = " ".join(format_string)

        # Count the number of people to generate
        people = format_string.count(IDENTIFIER)

        # Open the JSON file or create a new dict to load
        try:
            with open(FORMAT_FILE, 'r') as old:
                formats = json.load(old)
        except FileNotFoundError:
            formats = {}

        # Check if format string is in the dict
        try:
            formats[str(people)].remove(format_string)
        except KeyError:
            pass

        # Write the formats to the JSON file
        with open(FORMAT_FILE, 'w') as new:
            json.dump(formats, new)

        return "If format %s for %s people existed, it was removed!" % (code(format_string), code(str(people)))
