# Jobber [![Build Status](https://travis-ci.org/dshearer/jobber.svg?branch=master)](https://travis-ci.org/dshearer/jobber)

A replacement for cron, with sophisticated status-reporting and error-handling.

## Intro

Jobber is a lightweight utility for Unix-like systems that can run arbitrary commands, or "jobs", according to a schedule.  It is meant to be a replacement for the classic Unix utility [cron](http://en.wikipedia.org/wiki/Cron).

Along with the functionality of cron, Jobber also provides:
* **Job execution history**: you can see what jobs have recently run, and whether they succeeded or failed.
* **Sophisticated error handling**: you can control whether and when a job is run again after it fails.  For example, after an initial failure of a job, Jobber can schedule future runs using an exponential backoff algorithm.
* **Sophisticated error reporting**: you can control whether Jobber notifies you about each failed run, or only about jobs that have been disabled due to repeated failures.

## Contributing

**Contributions/suggestions/requests are welcome!**  Feel free to [open an issue](https://github.com/dshearer/jobber/issues), or ask a question on [the mailing list](https://groups.google.com/d/forum/jobber-proj).

## More Info

More info can be found on [Jobber's website](http://dshearer.github.io/jobber/).

## Debian / Ubuntu

HowTo documentation for Debian / Ubuntu based systems can be found in the [DEBIAN_UBUNTU.md](DEBIAN_UBUNTU.md) file.
