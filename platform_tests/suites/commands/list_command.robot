*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    List Command
Force Tags       test

*** Test Cases ***        ROOT_JOB_NAME    NON_ROOT_JOB_NAME    ROOT_LIST_RESULTS    ROOT_ALL_LIST_RESULTS    NON_ROOT_LIST_RESULTS
While Root Has Job        TestJob          ${EMPTY}             TestJob              TestJob                  ${EMPTY}
While Non-Root Has Job    ${EMPTY}         TestJob              ${EMPTY}             TestJob                  TestJob
While Both Have Jobs      TestJob1         TestJob2             TestJob1             TestJob1,TestJob2        TestJob2

*** Keywords ***
List Command
    [Arguments]    ${root_job_name}    ${non_root_job_name}
    ...            ${root_list_results}    ${root_list_all_results}
    ...            ${non_root_list_results}

    ${should_install_root_job}=       Evaluate    len('${root_job_name}') > 0
    ${should_install_non_root_job}=   Evaluate    len('${non_root_job_name}') > 0

    Run Keyword If    ${should_install_root_job}
    ...    Make and Install Jobfile    ${root_job_name}        for_root=${True}
    Run Keyword If    ${should_install_non_root_job}
    ...    Make and Install Jobfile    ${non_root_job_name}    for_root=${False}

    Jobber List Should Return    ${root_list_results}        as_root=${True}
    Jobber List Should Return    ${root_list_all_results}    as_root=${True}    all_users=${True}
    Jobber List Should Return    ${non_root_list_results}    as_root=${False}

Make and Install Jobfile
    [Arguments]    ${job_name}    ${for_root}

    ${jobfile}=    Make Jobfile    ${job_name}    exit 0
    Install Jobfile    ${jobfile}    for_root=${for_root}
    Nothing Has Crashed
