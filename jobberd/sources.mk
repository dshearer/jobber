DAEMON_SOURCES := \
	commands.go \
	context.go \
	ipc_server.go \
	jobberd.go \
	job_file.go \
	job.go \
	job_manager.go \
	job_runner_thread.go \
	logging.go \
	queue.go \
	run_rec_notifier.go \
	safe_bytes_to_str.go \
	sudo.go \
	sudo_cmd_linux.go \
	sudo_cmd_freebsd.go

DAEMON_TEST_SOURCES := \
	job_file_parse_test.go \
	next_run_time_test.go

