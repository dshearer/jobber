*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test

*** Test Cases ***
Random Time Spec
    # make jobfile with random time spec
    ${time_spec}=    Set Variable    0 0 R5-8
    ${jobfile}=    Make Jobfile    TestJob    exit 0    time=${time_spec}
    Install Jobfile    ${jobfile}    for_root=${True}

    # test
    Nothing Has Crashed
