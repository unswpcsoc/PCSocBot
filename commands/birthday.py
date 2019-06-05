import asyncio
from collections import defaultdict
import datetime
import json

from discord.utils import find

from commands.base import Command
from helpers import CommandFailure, bold
from configstartup import config


BIRTHDAY_FILE = config['FILES'].get('Birthday')
BDAY_ROLE_ID = config['ROLES'].get('Birthday')


class Birthday(Command):
    desc = "This command can be used to add or remove your birthday. When it is " \
        "your birthday, PCSocBot will give you the Birthday! role for a day."


class Add(Birthday):
    desc = "Add your own birthday. Please use the format dd/mm " \
        "(trailing zeroes aren't necessary)."

    def eval(self, birthday):
        dt_birthday = validate(birthday)
        if dt_birthday is None:
            raise CommandFailure("Please input a valid date format (dd/mm).")

        all_birthdays = get_birthdays(BIRTHDAY_FILE)

        # Check if they've already given their birthday
        curr_date = find_user(all_birthdays, self.user)
        if curr_date is not None:
            raise CommandFailure("You've already entered your birthday. "
                                 "If you wish to change it, please remove it first.")

        # Convert datetime object back to a consistent dd/mm string
        day_month = dt_birthday.strftime("%d/%m")

        # Add it and save
        all_birthdays[day_month].append(self.user)
        save_birthdays(all_birthdays)

        return f"{dt_birthday:%-d} {dt_birthday:%B} has been added as your birthday!"


class Remove(Birthday):
    desc = "Remove your birthday, and don't get the role on your birthday. " \
        "No arguments are needed."

    def eval(self):
        all_birthdays = get_birthdays(BIRTHDAY_FILE)

        # Check if they've given their birthday
        curr_date = find_user(all_birthdays, self.user)
        if curr_date is None:
            raise CommandFailure("You haven't supplied your birthday.")

        # Remove it and save
        all_birthdays[curr_date].remove(self.user)
        save_birthdays(all_birthdays)

        return "Your birthday has been removed."


class ModAdd(Birthday):
    desc = "Manually grant the Birthday! role to a given user and store " \
        "their birthday. Mod only."
    roles_required = ['mod', 'exec']

    async def eval(self, user):
        member = self.server.get_member_named(user)
        if member is None:
            raise CommandFailure(f"User {bold(user)} doesn't exist!")

        # Check if user already has a birthday added
        all_birthdays = get_birthdays(BIRTHDAY_FILE)
        curr_date = find_user(all_birthdays, member.id)
        dm_today = datetime.datetime.today().strftime("%d/%m")
        if curr_date is not None and curr_date != dm_today:
            raise CommandFailure(
                f"{bold(user)} already has a birthday set for "
                f"a different date - they don't need the date today")

        if curr_date is None:
            # User doesn't have any birthday set
            # Set their birthday to today for next year
            all_birthdays[dm_today] = member.id
            save_birthdays(all_birthdays)

        # Grant them the Birthday! role
        bday_role = get_role(self.server)
        await self.client.add_roles(member, bday_role)

        return f"{bold(user)} has been granted the Birthday role."


class ModPurge(Birthday):
    desc = "Remove the Birthday! role from all users. Mod only."
    roles_required = ['mod', 'exec']

    async def eval(self):
        s = self.server
        await remove_birthdays(self.client, s.members, get_role(s))
        return "Removed all Birthday! roles."


def get_birthdays(bday_file):
    """
    Gets JSON object of all birthdays.
    """
    all_birthdays = defaultdict(list)
    try:
        with open(bday_file) as birthdays:
            all_birthdays.update(json.load(birthdays))
    except FileNotFoundError:
        pass

    return all_birthdays


def save_birthdays(all_birthdays):
    """
    Dump the defaultdict as a JSON file.
    """
    with open(BIRTHDAY_FILE, "w") as birthdays:
        json.dump(all_birthdays, birthdays)


def validate(date_string):
    """
    Checks if a given string is a valid date.
    """
    try:
        return datetime.datetime.strptime(date_string, "%d/%m")
    except ValueError:
        return None


def find_user(birthdays, user):
    """
    Finds a user's birthday.
    Returns the day_month string if their birthday has been stored.
    Returns None if they haven't inputted their birthday.
    """
    for date, users in birthdays.items():
        if user in users:
            return date

    return None


def get_role(server):
    """
    Gets the Birthday! role in the server.
    """
    return find(lambda r: r.id == BDAY_ROLE_ID, server.roles)


async def remove_birthdays(client, members, bday_role):
    """
    Remove Birthday! role from all members in the server.
    """
    for member in members:
        if any(bday_role == role for role in member.roles):
            await client.remove_roles(member, bday_role)


async def update_birthday(client):
    """
    Update birthdays at the beginning of the day (00:00).
    """
    prev = datetime.datetime.today()
    while True:
        await asyncio.sleep(360)
        new = datetime.datetime.today()
        if new.day != prev.day:
            # It's a new day - remove all previous roles, add new roles
            all_birthdays = get_birthdays(BIRTHDAY_FILE)
            dm_today = new.strftime("%d/%m")

            # Get all members
            server = list(client.servers)[0]
            bday_role = get_role(server)

            # Remove everyone with the Birthday role from yesterday
            # TODO: Generalise this for any role, usable for any command
            await remove_birthdays(client, server.members, bday_role)

            # Happy Birthday!
            for birthday_member in all_birthdays[dm_today]:
                member = server.get_member(birthday_member)
                if member is not None:
                    await client.add_roles(member, bday_role)

        prev = new
