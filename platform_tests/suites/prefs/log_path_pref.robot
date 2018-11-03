*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown Log Path Pref
Force Tags       test    prefs    logpath

*** Test Cases ***
Log Path Preference
    ${log_path}=    Join Path    ~normuser    jobber-log
    Set Test Variable    ${log_path}
    File Should Not Exist    ${log_path}

    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob    exit 0    log_path=${log_path}
    Install Jobfile    ${jobfile}    for_root=${False}
    Nothing Has Crashed

    Sleep    3s    reason=Wait for job to run

    # test
    File Should Exist           ${log_path}    msg=Log file was not created
    File Should Not Be Empty    ${log_path}    msg=Log file is empty

*** Keywords ***
Teardown Log Path Pref
    Remove Files    ${log_path}
    Run Keyword    Teardown
