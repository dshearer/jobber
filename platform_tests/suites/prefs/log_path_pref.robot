*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test

*** Test Cases ***
Log Path Preference
    ${log_path}=    Set Variable    /home/normuser/.jobber-log
    File Should Not Exist    ${log_path}

    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob    exit 0
    Install Jobfile    ${jobfile}    for_root=${False}
    Nothing Has Crashed

    Sleep    3s    reason=Wait for job to run

    # test
    File Should Exist           ${log_path}    msg=Log file was not created
    File Should Not Be Empty    ${log_path}    msg=Log file is empty
