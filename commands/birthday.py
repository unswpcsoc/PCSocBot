import asyncio
from collections import defaultdict
import datetime
import json
import re

from commands.base import Command
from helpers import CommandFailure

BIRTHDAY_FILE = "files/birthdays.json"
BIRTHDAY_ROLE_NAME = "Birthday!"

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
        all_birthdays[day_month].append(self.user)

        with open(BIRTHDAY_FILE, "w") as birthdays:
            json.dump(all_birthdays, birthdays)

        return "Your birthday has been added!"


class Remove(Birthday):
    desc = "Remove your birthday, and don't get the role on your birthday. " \
        "No arguments are needed."

    def eval(self):
        all_birthdays = get_birthdays(BIRTHDAY_FILE)

        # Check if they've given their birthday
        curr_date = find_user(all_birthdays, self.user)
        if curr_date is None:
            raise CommandFailure("You haven't supplied your birthday.")

        all_birthdays[curr_date].remove(self.user)
        with open(BIRTHDAY_FILE, "w") as birthdays:
            json.dump(all_birthdays, birthdays)

        return "Your birthday has been removed."



def get_birthdays(bday_file):
    """
    Gets JSON object of all birthdays
    """
    all_birthdays = defaultdict(list)
    try:
        with open(bday_file) as birthdays:
            all_birthdays.update(json.load(birthdays))
    except FileNotFoundError:
        pass

    return all_birthdays


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
    # use for ... checking if they've already added and for removing
    for date, users in birthdays.items():
        if user in users:
            return date

    return None


async def update_birthday(client):
    """
    Update birthdays at the beginning of the day (00:00).
    """
    prev = datetime.datetime.today()
    while True:
        await asyncio.sleep(5)
        new = datetime.datetime.today()
        if new.day != prev.day:
            # It's a new day - remove all previous roles, add new roles
            all_birthdays = get_birthdays(BIRTHDAY_FILE)
            dm_today = new.strftime("%d/%m")
            
            # Get all members
            server = list(client.servers)[0]
            members = server.members
            birthday_role = next(x for x in server.roles if x.name == BIRTHDAY_ROLE_NAME)

            # Remove everyone with the Birthday role from yesterday
            for member in members:
                if any(birthday_role == role for role in member.roles):
                    await client.remove_roles(member, birthday_role)

            # Happy Birthday!
            for birthday_member in all_birthdays[dm_today]:
                member = server.get_member(birthday_member)
                if member is not None:
                    await client.add_roles(member, birthday_role)

        prev = new