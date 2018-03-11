JOBFILE_SOURCES := \
	jobfile/error_handler.go \
	jobfile/file_run_log.go \
	jobfile/job_file.go \
	jobfile/job_output_handler.go \
	jobfile/job.go \
	jobfile/mem_only_run_log.go \
	jobfile/parse_time_spec.y \
	jobfile/run_log.go \
	jobfile/run_rec_notifier.go \
	jobfile/safe_bytes_to_str.go \
	jobfile/sources.mk \
	jobfile/time_spec.go

JOBFILE_TEST_SOURCES := \
	jobfile/file_run_log_test.go \
	jobfile/job_file_parse_test.go \
	jobfile/run_log_test.go
