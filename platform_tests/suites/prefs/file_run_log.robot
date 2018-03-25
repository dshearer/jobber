*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    File Run Log
Force Tags       test    prefs

*** Test Cases ***    AS_ROOT
As Root               ${True}
As Non-Root           ${False}

*** Keywords ***
File Run Log
    [Arguments]    ${as_root}

    ${run_log_path}=    Make Tempfile
    Make and Install Jobfile    ${as_root}    ${run_log_path}

    Sleep    3s    reason=Give jobs time to run

    # check
    File Should Not Be Empty    ${run_log_path}
    Jobber Log Should Return Something    ${as_root}
    Nothing Has Crashed

Make and Install Jobfile
    [Arguments]    ${for_root}    ${run_log_path}

    ${jobfile}=    Make Jobfile    TestJob    exit 0    file_run_log_path=${run_log_path}
    Install Jobfile    ${jobfile}    for_root=${for_root}
    Nothing Has Crashed

Jobber Log Should Return Something
    [Arguments]    ${as_root}

    ${num_entries}=    Jobber Log    as_root=${as_root}    all_users=${False}
    Should Be True    ${num_entries} > 0
