*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    config

*** Test Cases ***
Default Config Is Installed
    Config File Should Exist

Prefs File Excludes User
    ${user}=    Set Variable    normuser
    Jobberrunner Should Be Running For User    ${user}
    Make Config That Excludes User    ${user}
    Restart Jobber Service
    Jobberrunner Should Not Be Running For User    ${user}

*** Keywords ***
Make Config That Excludes User
    [Arguments]    ${user}
    Set Config    exclude_users=${user}

Restart Jobber Service
    # restart jobber
    Restart Service
    Nothing Has Crashed

    # so teardown doesn't fail
    ${runner_procs}=    Runner Proc Info
    Set Test Variable    ${runner_procs}
