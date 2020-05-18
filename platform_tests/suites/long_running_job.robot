*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test   long

*** Test Cases ***
Reload Should Work While Job Is Running
    # make jobfile
    ${cmd}=    Set Variable    sleep 600
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    Install Jobfile    ${jobfile}    for_root=${True}

    # wait, then reload
    Sleep    3s    reason=Wait for job to run
    ${reload_succeeded}=    Jobber Try Reload

    # test
    Nothing Has Crashed
    Should Be True  ${reload_succeeded}     msg=Reload cmd failed
