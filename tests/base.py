import re

import commands


class GenericObject(object):
    # object has no way to set attributes for it
    pass


class Message:
    def __init__(self, content, **kwargs):
        self.content = commands.PREFIX + content
        self.author = GenericObject()
        self.author.id = int(kwargs.pop('id', 1))
        self.author.name = kwargs.pop('name', 'matt')
        self.server = GenericObject()
        self.server.members = None


def run_command(command, **kwargs):
    message = Message(command, **kwargs)
    command = command.split()
    cls, args = commands.Help.find_command(command)
    return cls(None, message, *args).output

kwargs = {}
while True:
    command = input("Enter a command: ")
    result = re.match(r'([a-z]+)=(\w+)', command)
    if result:
        kwargs[result.group(1)] = result.group(2)
        print("set", result.group(1), "to", result.group(2))
    else:
        print(run_command(command, **kwargs))