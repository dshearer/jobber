# Jobber

A replacement for cron, with sophisticated status-reporting and error-handling.

# Intro

Jobber is a utility for Unix-like systems that can run arbitrary commands, or "jobs", according to a schedule.  It is meant to be a replacement for the classic Unix utility [cron](http://en.wikipedia.org/wiki/Cron).

Along with the functionality of cron, Jobber also provides:
* **Job execution history**: you can see what jobs have recently run, and whether they succeeded or failed.
* **Sophisticated error handling**: you can control whether and when a job is run again after it fails.  For example, after an initial failure of a job, Jobber can schedule future runs using an exponential backoff algorithm.
* **Sophisticated error reporting**: you can control whether Jobber notifies you about each failed run, or only about jobs that have been disabled due to repeated failures.

# Maturity

Jobber is curerntly at a "beta" level of maturity.  The largest open task is to do a thorough security review.

# Target Systems

Jobber is written in [Go](http://golang.org/), and it should be possible to compile and run it on any modern Unix-like system.  However, at this time its installation script targets RHEL, Fedora, and CentOS.  (Actually, it has only been tested on CentOS....)

# Installation

You need a recent version of [Go](http://golang.org/) to compile Jobber.  Once you've installed Go and set up your Go workspace, you can build it thus:

    cd /path/to/your/workspace
    go get github.com/dshearer/jobber
    make -C src/github.com/dshearer/jobber

As mentioned above, Jobber's installation script currently targets RHEL-like systems.  If you are on such a system, you can install Jobber thus:

    cd /path/to/your/workspace
    sudo make install

# Usage

## Defining Jobs

As with cron, each user can have its own set of jobs, which will be run under that user's privileges.  A user's jobs are defined in a file named ".jobber" in the user's home directory.  Jobfiles are written in JSON format.  Here's an example:

    [
        {
            "name": "DailyBackup",
            "cmd": "backup daily",
            "time": {
                "sec": 0,
                "min": 0,
                "hour": 13
            },
            "onError": "Stop",
            "notifyOnError": false,
            "notifyOnFailure": true
        },
        {
            "name": "WeeklyBackup",
            "cmd": "backup weekly",
            "time": {
                "sec": 0,
                "min": 0,
                "hour": 14,
                "wday": 1
            },
            "onError": "Stop",
            "notifyOnError": false,
            "notifyOnFailure": true
        }
    ]

This jobfile defines two jobs.  Field "name" is self-explanatory.  Field "cmd" can contain any shell command.

### Field "time"

Field "time" specifies when the job is run, in a manner similar to how cron jobs' schedules are specified.  A job is scheduled thus: at each second, jobber looks at the job's "time" field and determines whether all of the subfields that happen to be present match the current time; if so it runs the job, and otherwise it does not.  Thus, "DailyBackup" is run whenever the current time is 13:00:00 (i.e., 1 PM), no matter the day, whereas "WeeklyBackup" is run whenever the current time is 14:00:00 (i.e., 2 PM) and the current weekday is Monday.  The possible subfields of "time" and their value ranges are as follows:

* "sec": 0 &ndash; 59
* "min": 0 &ndash; 59
* "hour": 0 &ndash; 23
* "mday": 1 &ndash; 31
* "mon": 1 &ndash; 12
* "wday": 1 &ndash; 7

### Fields "onError", "notifyOnError", and "notifyOnFailure"

When discussnig jobs, by "job error" we mean the failure of a particular run of a job, whereas by "job failure" we mean a job that jobber has decided not to schedule anymore due to one or more recent job errors.  Jobber considers a run to have failed when the command (in field "cmd") returns a non-0 exit status.

Field "onError" specifies what jobber will do when a job error occurs.  The possible values:

* "Stop": *Stop scheduling runs of this job.*  That is, a single job error results in a job failure.
* "Backoff": *Schedule runs of this job according to an exponential backoff algorithm.*  If a later run of the job succeeds, jobber resumes scheduling this job normally; but if the job fails again on several consecutive runs, jobber stops scheduling it &emdash; that is, the job fails.
* "Continue": *Continue scheduling this job normally.*  That is, the job will never fail.

Fields "notifyOnError" and "notifyOnFailure" control whether the user that owns a job is notified about job errors and job failures.

## Loading Jobs

After you've created a user's jobfile, log in as that user and do:

    jobber reload

You can also reload all users' jobfiles by logging in as root and doing:

    jobber reload -a

## Listing Jobs

You can list the jobs for a particular user by logging in as that user and doing

    jobber list

This command also shows you the status of each job &mdash; that is, whether the job is being scheduled as normal, the exponential backoff algorithm is being applied, or the job has failed.

As with the "reload" command, you can do the same for all users by adding the "-a" option as root.

## Listing Runs

You can see a list of recent runs of any jobs for a particular user by logging in as that user and doing

    jobber log

As with the other commands, you can do the same for all users by adding the "-a" option as root.

## Testing Jobs

If you'd like to test out a job, do

    jobber test JOB_NAME

Jobber will immediately run that job, tell you whether it succeeded, and show you its output.
