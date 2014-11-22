jobber
======

A replacement for cron, with sophisticated status-reporting and error-handling.

Intro
-----
Jobber is a utility for Unix-like systems that can run arbitrary commands, or "jobs", according to a schedule.  It is meant to be a replacement for the classic Unix utility [cron](http://en.wikipedia.org/wiki/Cron).

Along with the functionality of cron, Jobber also provides:
* **Job execution history**: you can see what jobs have recently run, and whether they succeeded or failed.
* **Sophisticated error handling**: you can control whether and when a job is run again after it fails.  For example, after an initial failure of a job, Jobber can schedule future runs using an exponential backoff algorithm.
* **Sophisticated error reporting**: you can control whether Jobber notifies you about each failed run, or only about jobs that have been disabled due to repeated failures.

Target Systems
--------------
Jobber is written in [Go](http://golang.org/), and it should be possible to compile and run it on any modern Unix-like system.  However, at this time its installation script targets RHEL, Fedora, and CentOS.  (Actually, it has only been tested on CentOS....)

Installation
------------
You need a recent version of [Go](http://golang.org/) to compile Jobber.  Once you've installed Go and set up your Go workspace, you can build it thus:

    cd /path/to/your/workspace
    go get github.com/dshearer/jobber
    make -C src/github.com/dshearer/jobber

As mentioned above, Jobber's installation script currently targets RHEL-like systems.  If you are on such a system, you can install Jobber thus:

    cd /path/to/your/workspace
    sudo make install
