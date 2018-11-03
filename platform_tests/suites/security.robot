*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    security

*** Test Cases ***
Privilege Separation
    # make jobfile for normal user
    ${output_file}=    Make Tempfile    create=${True}
    ${cmd}=    Set Variable    echo 'Hello' > ${output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}

    # change owner and mode of output file
    Chown    ${output_file}    root
    Chmod    ${output_file}    0600

    # install jobfile
    Install Jobfile    ${jobfile}    for_root=${False}
    Nothing Has Crashed

    Sleep    3s    reason=Wait for job to run

    # test
    Nothing Has Crashed
    File Should Be Empty    ${output_file}
