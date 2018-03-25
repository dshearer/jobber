*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    Job Output Is Written to Disk (or Not)
Force Tags       test    prefs    joboutput

*** Variables ***
${JOB_NAME}             TestJob
${STDOUT_EXP_OUTPUT}    Hi from stdout
${STDERR_EXP_OUTPUT}    Hi from stderr

*** Test Cases ***                 STDOUT     STDERR
Write Stdout to Disk               ${True}    ${False}
Write Stderr to Disk               ${False}   ${True}
Write Stdout and Stderr to Disk    ${True}    ${True}
Write Neither to Disk              ${False}   ${False}

*** Keywords ***
Job Output Is Written to Disk (or Not)
    [Arguments]    ${stdout}    ${stderr}

    # install jobfile
    ${output_dir}=    Install Jobfile with Job that Produces Output    stdout=${stdout}   stderr=${stderr}

    Sleep    3s    reason=Wait for job to run

    # check stdout output files
    Run Keyword If    ${stdout}
    ...            Stdout Output Files Should Exist with Expected Contents    ${output_dir}
    ...    ELSE    Stdout Output Files Should Not Exist    ${output_dir}

    # check stderr output files
    Run Keyword If    ${stderr}
    ...            Stderr Output Files Should Exist with Expected Contents    ${output_dir}
    ...    ELSE    Stderr Output Files Should Not Exist    ${output_dir}

Install Jobfile with Job that Produces Output
    [Arguments]    ${stdout}    ${stderr}

    # make command
    ${cmd}=    Set Variable    echo -n '${STDOUT_EXP_OUTPUT}'\necho -n '${STDERR_EXP_OUTPUT}' >&2

    # make jobfile
    ${output_dir}=    Make Tempfile
    Create Directory    ${output_dir}
    ${jobfile}=    Run Keyword If    ${stdout} and ${stderr}
    ...                            Make Jobfile    job_name=${JOB_NAME}    cmd=${cmd}
    ...                            notify_on_success=${True}    stdout_output_dir=${output_dir}
    ...                            stdout_output_max_age=1      stderr_output_dir=${output_dir}
    ...                            stderr_output_max_age=1
    ...    ELSE IF    ${stdout}    Make Jobfile    job_name=${JOB_NAME}    cmd=${cmd}
    ...                            notify_on_success=${True}    stdout_output_dir=${output_dir}
    ...                            stdout_output_max_age=1
    ...    ELSE IF    ${stderr}    Make Jobfile    job_name=${JOB_NAME}    cmd=${cmd}
    ...                            notify_on_success=${True}    stderr_output_dir=${output_dir}
    ...                            stderr_output_max_age=1
    ...    ELSE                    Make Jobfile    job_name=${JOB_NAME}    cmd=${cmd}
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed
    [Return]    ${output_dir}

Stdout Output Files Should Exist with Expected Contents
    [Arguments]    ${output_dir}

    @{output_files}=    Output Files Should Exist        ${output_dir}    stdout=${True}
    Output Files Should Have Contents    ${output_files}    ${STDOUT_EXP_OUTPUT}

Stdout Output Files Should Not Exist
    [Arguments]    ${output_dir}

    Output Files Should Not Exist    ${output_dir}    stdout=${True}

Stderr Output Files Should Exist with Expected Contents
    [Arguments]    ${output_dir}

    @{output_files}=    Output Files Should Exist        ${output_dir}    stdout=${False}
    Output Files Should Have Contents    ${output_files}    ${STDERR_EXP_OUTPUT}

Stderr Output Files Should Not Exist
    [Arguments]    ${output_dir}

    Output Files Should Not Exist    ${output_dir}    stdout=${False}

Output Files Should Exist
    [Arguments]    ${output_dir}    ${stdout}

    ${dir_path}=    Set Variable    ${output_dir}/${JOB_NAME}
    Directory Should Exist    ${dir_path}
    ${pattern}=    Run Keyword If    ${stdout}
    ...            Set Variable    *.stdout
    ...    ELSE    Set Variable    *.stderr
    @{output_files}=    List Files in Directory    ${dir_path}    ${pattern}    absolute=${True}
    Should Not Be Empty    ${output_files}
    [return]    @{output_files}

Output Files Should Not Exist
    [Arguments]    ${output_dir}    ${stdout}

    ${dir_path}=    Set Variable    ${output_dir}/${JOB_NAME}
    ${dir_exists}=    testlib.Directory Exists    ${dir_path}
    ${empty_list}=    Evaluate    []
    Return From Keyword If    not ${dir_exists}    ${empty_list}

    ${pattern}=    Run Keyword If    ${stdout}
    ...            Set Variable    *.stdout
    ...    ELSE    Set Variable    *.stderr
    @{output_files}=    List Files in Directory    ${dir_path}    ${pattern}    absolute=${True}
    Should Be Empty    ${output_files}
    [return]    @{output_files}

Output Files Should Have Contents
    [Arguments]    ${output_files}    ${expected_output}

    :FOR    ${file}    IN    @{output_files}
    \    File Should Have Contents    ${file}    ${expected_output}
