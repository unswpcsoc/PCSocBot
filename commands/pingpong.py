from commands.base import Command


class Ping(Command):
    desc = "Pong! This command could be used to check if the bot is up. Or entertainment when you're bored."

    def eval(self):
        return "Pong!"


class Pong(Command):
    desc = "Ping! This command could be used to check if the bot is up. Or entertainment when you're bored."

    def eval(self):
        return "Ping!"
