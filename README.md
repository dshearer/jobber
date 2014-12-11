# Jobber

A replacement for cron, with sophisticated status-reporting and error-handling.

# Intro

Jobber is a lightweight utility for Unix-like systems that can run arbitrary commands, or "jobs", according to a schedule.  It is meant to be a replacement for the classic Unix utility [cron](http://en.wikipedia.org/wiki/Cron).

Along with the functionality of cron, Jobber also provides:
* **Job execution history**: you can see what jobs have recently run, and whether they succeeded or failed.
* **Sophisticated error handling**: you can control whether and when a job is run again after it fails.  For example, after an initial failure of a job, Jobber can schedule future runs using an exponential backoff algorithm.
* **Sophisticated error reporting**: you can control whether Jobber notifies you about each failed run, or only about jobs that have been disabled due to repeated failures.

# More Info

More info can be found on [Jobber's website](http://dshearer.github.io/jobber/).
