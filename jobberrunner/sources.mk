RUNNER_SOURCES := \
	jobberrunner/cmd_cat.go \
	jobberrunner/cmd_delete_job.go \
	jobberrunner/cmd_init.go \
	jobberrunner/cmd_list_jobs.go \
	jobberrunner/cmd_log.go \
	jobberrunner/cmd_pause.go \
	jobberrunner/cmd_reload.go \
	jobberrunner/cmd_resume.go \
	jobberrunner/cmd_set_job.go \
	jobberrunner/cmd_test_job.go \
	jobberrunner/ipc_server.go \
	jobberrunner/job_manager.go \
	jobberrunner/job_runner_thread.go \
	jobberrunner/main.go \
	jobberrunner/queue.go \
	jobberrunner/sources.mk \
	jobberrunner/testjob/test_job_server.go \
	jobberrunner/testjob/test_job_thread.go \

RUNNER_TEST_SOURCES := \
	jobberrunner/cmd_init_test.go \
	jobberrunner/next_run_time_test.go
