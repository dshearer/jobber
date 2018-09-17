*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    Log Command
Force Tags       test    cmd

*** Test Cases ***          ROOT_HAS_JOB    NON_ROOT_HAS_JOB    ROOT_LOG_RESULTS    ROOT_LOG_ALL_RESULTS    NON_ROOT_LOG_RESULTS
While Root Has Job          ${True}         ${False}            ${True}             ${True}                 ${False}
While Non-Root Has Job      ${False}        ${True}             ${False}            ${True}                 ${True}
While Both Have Job         ${True}         ${True}             ${True}             ${True}                 ${True}
While Neither Have Job      ${False}        ${False}            ${False}            ${False}                ${False}

*** Keywords ***
Log Command
    [Arguments]    ${root_has_job}    ${non_root_has_job}    ${root_log_results}
    ...    ${root_log_all_results}    ${non_root_log_results}

    # make jobfiles
    Run Keyword If    ${root_has_job}        Make and Install Jobfile    ${True}
    Run Keyword If    ${non_root_has_job}    Make and Install Jobfile    ${False}

    Sleep    3s    reason=Give jobs time to run

    # test
    Run Keyword If    ${root_log_results}
    ...            Jobber Log Should Return Something        as_root=${True}    all_users=${False}
    ...    ELSE    Jobber Log Should Not Return Something    as_root=${True}    all_users=${False}
    Run Keyword If    ${root_log_all_results}
    ...            Jobber Log Should Return Something        as_root=${True}    all_users=${True}
    ...    ELSE    Jobber Log Should Not Return Something    as_root=${True}    all_users=${True}
    Run Keyword If    ${non_root_log_results}
    ...            Jobber Log Should Return Something        as_root=${False}   all_users=${False}
    ...    ELSE    Jobber Log Should Not Return Something    as_root=${False}   all_users=${False}

Make and Install Jobfile
    [Arguments]    ${for_root}

    ${jobfile}=    Make Jobfile    TestJob    exit 0
    Install Jobfile    ${jobfile}    for_root=${for_root}
    Nothing Has Crashed

Jobber Log Should Return Something
    [Arguments]    ${as_root}    ${all_users}

    ${num_entries}=    Jobber Log    as_root=${as_root}    all_users=${all_users}
    Should Be True    ${num_entries} > 0

Jobber Log Should Not Return Something
    [Arguments]    ${as_root}    ${all_users}

    ${num_entries}=    Jobber Log    as_root=${as_root}    all_users=${all_users}
    Should Be True    ${num_entries} == 0
