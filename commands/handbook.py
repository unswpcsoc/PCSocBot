from commands.base import Command
from discord import Embed
from helpers import bold, CommandFailure, is_good_response

import bs4
import re
import requests

HANDBOOK_URL = 'https://www.handbook.unsw.edu.au/{}/courses/2019/{}'


def subject_details(code):
    code = code.upper()
    url = HANDBOOK_URL.format('undergraduate', code)
    page = requests.get(url)

    if not is_good_response(page):
        # Check postgrad
        url = HANDBOOK_URL.format('postgraduate', code)
        page = requests.get(url)
        if not is_good_response(page):
            return None

    soup = bs4.BeautifulSoup(page.text, features='lxml')
    course_title = soup.find(
        'span', attrs={'data-hbui': 'module-title'}).string.strip()
    course_offerings = soup.find('strong', string=re.compile(
        'Offering Terms')).parent.contents[3].string.strip()
    course_conditions_tag = soup.find('div', id='readMoreSubjectConditions')

    course_conditions = ''
    if course_conditions_tag is not None:
        for s in course_conditions_tag.contents[1].contents[1].strings:
            course_conditions += s
        course_conditions = course_conditions.strip()
    else:
        course_conditions = 'None'

    course_desc = ''
    # Get first div
    course_desc_cont = soup.select_one('#readMoreIntro').find('div')
    if course_desc_cont is not None:
        course_desc = course_desc_cont.text.strip()
        # Just get the first paragraph
        course_desc = course_desc.split('\n', 1)[0]

    return {
        'title': course_title,
        'description': course_desc,
        'offerings': course_offerings,
        'conditions': course_conditions,
        'link': url
    }


class Handbook(Command):
    desc = "This command scrapes entries in the UNSW handbook."

    def eval(self, course_code):
        if re.search(r'^[a-zA-Z]{4}[0-9]{4}$', course_code) is None:
            raise CommandFailure(
                f'Incorrectly formatted course code: {bold(course_code)}')

        course = subject_details(course_code)
        if course is None:
            return f'Course {course_code} could not be found'

        ret = Embed(
            title=course['title'],
            description=course['description'],
            url=course['link'],
            color=self.EMBED_COLOR
        )
        ret.add_field(name='Offering Terms', value=course['offerings'])
        ret.add_field(name='Enrolment Conditions', value=course['conditions'])
        return ret
