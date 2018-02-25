*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    Jobber Should Run Job in Jobfile
Force Tags     test

*** Test Cases ***    AS_ROOT
As Root               ${True}
As Non-Root           ${False}

*** Keywords ***
Jobber Should Run Job in Jobfile
    [Arguments]    ${as_root}

    ${expected_output}=    Set Variable    Hello

    # make jobfile
    ${output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n '${expected_output}' > ${output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    Install Jobfile    ${jobfile}    for_root=${as_root}

    # wait
    Sleep    3s    reason=Wait for job to run

    # test
    Nothing Has Crashed
    File Should Have Contents    ${output_file}    ${expected_output}    msg=job didn't run
