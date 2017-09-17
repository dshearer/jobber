#
# Regular cron jobs for the jobber package
#
0 4	* * *	root	[ -x /usr/bin/jobber_maintenance ] && /usr/bin/jobber_maintenance
