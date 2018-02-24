import requests
import sys

def main():
    url = 'https://api.github.com/repos/dshearer/jobber/releases/latest'
    headers = {'Accept': 'application/vnd.github.v3+json'}
    resp = requests.get(url, headers=headers)
    resp.raise_for_status()
    sys.stdout.buffer.write(resp.content)

if __name__ == '__main__':
    main()
