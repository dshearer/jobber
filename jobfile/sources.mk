JOBFILE_SOURCES := \
	jobfile/error_handler.go \
	jobfile/file_run_log.go \
	jobfile/job_file.go \
	jobfile/job_output_handler.go \
	jobfile/job.go \
	jobfile/mem_only_run_log.go \
	jobfile/parse_time_spec.y \
	jobfile/result_sink_filesystem.go \
	jobfile/result_sink_program.go \
	jobfile/result_sink_socket.go \
	jobfile/result_sink_stdout.go \
	jobfile/result_sink_system_email.go \
	jobfile/result_sink.go \
	jobfile/run_log.go \
	jobfile/run_rec_server.go \
	jobfile/safe_bytes_to_str.go \
	jobfile/semver.go \
	jobfile/sources.mk \
	jobfile/time_spec.go

JOBFILE_TEST_SOURCES := \
	jobfile/file_run_log_test.go \
	jobfile/job_file_v1v2_parse_test.go \
	jobfile/job_file_v3_parse_test.go \
	jobfile/parse_time_spec_test.go \
	jobfile/run_log_test.go
