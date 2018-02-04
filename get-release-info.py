import requests
import sys

def reformat_date(date):
    dt = datetime.strptime(date, INPUT_DATE_FORMAT)
    return dt.strftime(OUTPUT_DATE_FORMAT)

def main():
    url = 'https://api.github.com/repos/dshearer/jobber/releases/latest'
    headers = {'Accept': 'application/vnd.github.v3+json'}
    resp = requests.get(url, headers=headers)
    resp.raise_for_status()
    print(resp.content)

if __name__ == '__main__':
    main()