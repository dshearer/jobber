RUNNER_SOURCES := \
	cmd_cat.go \
	cmd_init.go \
	cmd_list_jobs.go \
	cmd_log.go \
	cmd_pause.go \
	cmd_reload.go \
	cmd_resume.go \
	cmd_test_job.go \
	ipc_server.go \
	job_manager.go \
	job_runner_thread.go \
	main.go \
	queue.go

RUNNER_TEST_SOURCES := \
	cmd_init_test.go \
	next_run_time_test.go