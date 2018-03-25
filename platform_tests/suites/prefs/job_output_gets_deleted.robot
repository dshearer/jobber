*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    prefs    joboutput

*** Variables ***
${JOB_NAME}                   TestJob
${JOB_OUTPUT_MAX_AGE_DAYS}    5

*** Test Cases ***
Job Output Is Deleted on Schedule
    # make jobfile
    ${output_dir}=    Install Jobfile with Job that Produces Output

    Sleep    3s    reason=Wait for job to run
    Pause Job    ${JOB_NAME}

    # push the output files' timestamps back
    ${push_back_days}=    Evaluate    ${JOB_OUTPUT_MAX_AGE_DAYS} + 1
    testlib.Decrease Job Output Files Timestamps    ${output_dir}/${JOB_NAME}    ${push_back_days}
    Old Output Files Should Exist    ${output_dir}

    Resume Job    ${JOB_NAME}

    # check
    Wait Until Keyword Succeeds    3s    1s    Old Output Files Should Not Exist    ${output_dir}

*** Keywords ***
Install Jobfile with Job that Produces Output
    ${output_dir}=    Make Tempfile
    Create Directory    ${output_dir}
    ${cmd}=    Set Variable    echo -n 'Hi'
    ${jobfile}=    Make Jobfile    job_name=${JOB_NAME}    cmd=${cmd}
    ...    notify_on_success=${True}    stdout_output_dir=${output_dir}
    ...    stdout_output_max_age=${JOB_OUTPUT_MAX_AGE_DAYS}
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed
    [Return]    ${output_dir}

Old Output Files Should Exist
    [Arguments]    ${output_dir}

    ${actual_max_age}=    Max Job Output File Age    ${output_dir}/${JOB_NAME}
    ${gt}=    Evaluate    ${actual_max_age} > ${JOB_OUTPUT_MAX_AGE_DAYS}
    Should Be True    ${gt}    message=Old output files do not exist

Old Output Files Should Not Exist
    [Arguments]    ${output_dir}

    ${actual_max_age}=    Max Job Output File Age    ${output_dir}/${JOB_NAME}
    ${lt}=    Evaluate    ${actual_max_age} < ${JOB_OUTPUT_MAX_AGE_DAYS}
    Should Be True    ${lt}    message=Old output files still exist
