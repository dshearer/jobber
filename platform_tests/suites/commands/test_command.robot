*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test

*** Test Cases ***
Test Command
    # make & install jobfile
    ${output_file}=    Make Tempfile
    Make and Install Jobfile    ${output_file}

    # pause it
    Pause Job    TestJob
    Nothing Has Crashed

    # It's possible that the job is running at this very moment, so
    # wait a second to let it finish.
    Sleep    1s

    # get latest output (if any)
    ${output_1}=    Get File    ${output_file}

    # call 'jobber test'
    Test Job    TestJob
    Nothing Has Crashed

    # check whether it ran again
    ${output_2}=    Get File    ${output_file}
    Should Not Be Equal    ${output_1}    ${output_2}    msg=Job did not run

*** Keywords ***
Make and Install Jobfile
    [Arguments]    ${output_file}

    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed
