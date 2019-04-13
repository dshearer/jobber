import requests
import sys
import argparse
import json

URL = 'https://api.github.com/repos/dshearer/jobber/issues?state=closed&direction=asc&labels={label}'

def get_issues(milestone, label):
    url = URL.format(label=label)
    while True:
        resp = requests.get(url)
        resp.raise_for_status()
        for issue in json.loads(resp.content):
            ms = issue.get('milestone')
            if ms is None or ms['title'] != milestone:
                continue
            yield issue

        # go to next page
        links = resp.headers.get('Link')
        if links is None:
            return
        links = links.strip().split(',')
        next_link = None
        for link in links:
            if 'rel="next"' in link:
                next_link = link
        if next_link is None:
            return
        parts = next_link.split(';')
        url = parts[0].strip('<>')

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('MILESTONE')
    args = parser.parse_args()

    print("Enhancements:")
    for issue in get_issues(args.MILESTONE, 'enhancement'):
        print("* (#{}) {}".format(issue['number'], issue['title']))

    print("\nBugfixes:")
    for issue in get_issues(args.MILESTONE, 'bug'):
        print("* (#{}) {}".format(issue['number'], issue['title']))

if __name__ == '__main__':
    main()
