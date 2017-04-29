## Install Go on your Debian / Ubuntu based system

### Install build dependencies
```
$ apt-get update && apt-get install -y autoconf build-essential imagemagick libbz2-dev libcurl4-openssl-dev libevent-dev libffi-dev libglib2.0-dev libjpeg-dev libmagickcore-dev libmagickwand-dev libmysqlclient-dev libncurses-dev libpq-dev libreadline-dev libsqlite3-dev libssl-dev libxml2-dev libxslt-dev libyaml-dev zlib1g-dev git curl
```

### Install Go

Download the latest version of GoLang from https://golang.org/dl/
```
$ cd /tmp
$ curl -L https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz > go1.5.3.linux-amd64.tar.gz
$ tar -C /usr/local -xzvf go1.5.3.linux-amd64.tar.gz
```

Add Go to your $PATH variale by adding `:/usr/local/go/bin` to the environment path
```
$ vi /etc/environment
PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/usr/local/go/bin"
```

Create the directory where your GoLang path will be
```
$ mkdir -p /opt/go
$ chmod 775 /opt/go
$ chown root:staff /opt/go
```

Edit the `.bashrc` file of the user that is using GoLang. Add the following lines to that file.
```
export GOPATH=/opt/go
```

## Compile Jobber
```
$ go get github.com/dshearer/jobber
can't load package: package github.com/dshearer/jobber: no buildable Go source files in /opt/go/src/github.com/dshearer/jobber
$ cd /opt/go/src/github.com/dshearer/jobber
$ make
```

## Configure Jobber

Create a new, non-login user called “jobber_client”:

```
$ useradd --home / -M --system --shell /sbin/nologin jobber_client
```

## Install the binaries

Now, copy jobber to `/usr/local/bin` resp `/usr/local/sbin` and set its owner to jobber_client:root and its permissions to 4755. Copy jobberd to wherever and set its owner to root:root and its permissions to 0755.

```
$ cd /opt/go/bin/
$ cp jobber /usr/local/bin/.
$ cp jobberd /usr/local/sbin/.
$ cd /usr/local/bin
$ chown jobber_client:root jobber
$ chmod 4775 jobber
$ cd /usr/local/sbin
$ chown root:root jobberd
$ chmod 0755 jobberd
```

## Run Jobberd in the backend / Run Go Applications in the background

There's a huge number of options here, but we'll look at a stable, popular and cross-distro approach called Supervisor. Supervisor is a process management tool that handles restarting, recovering and managing logs, without requiring anything from your application (i.e. no PID files!).

### Install Supervisor
```
$ apt-get install supervisor
```

A simple configuration for our script, saved at /etc/supervisor/conf.d/jobberd.conf, would look like so:

```
[program:jobberd]
command=/usr/local/sbin/jobberd
autostart=true
autorestart=true
stderr_logfile=/var/log/jobber.err.log
stdout_logfile=/var/log/jobber.out.log
```

Once our configuration file is created and saved, we can inform Supervisor of our new program through the supervisorctl command. First we tell Supervisor to look for any new or changed program configurations in the /etc/supervisor/conf.d directory with:

```
$ supervisorctl reread
```

Followed by telling it to enact any changes with:

```
$ supervisorctl update
```
Any time you make a change to any program configuration file, running the two previous commands will bring the changes into effect.

Verify that Jobber is active
```
$ ps -ax | grep jobber
```

### More information

* https://www.digitalocean.com/community/tutorials/how-to-install-and-manage-supervisor-on-ubuntu-and-debian-vps

## Use Jobber

Login as a normal user
```
$ vi .jobber
```

```yaml
---
- name: Trigger FeedManager
  cmd: curl -u user:pass -i -H 'Accept:application/json'  http://app.example.com/trigger
  time: "0 0 * * * *"
  onError: Stop
  notifyOnError: true
  notifyOnFailure: true
```

Reload jobs

```
$ jobber reload
```

List jobs

```
$ jobber list
```

Listing runs

```
$ jobber log
```

Testing Jobs

```
$ jobber test JOB_NAME
```
