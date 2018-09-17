*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Test Template    Notify Pref Should Work
Force Tags     test    prefs

*** Test Cases ***                        ON_SUCCESS    ON_ERROR    JOB_SUCCEEDS    SHOULD_HAVE_CALLED_NOTIFY_PROGRAM
Notify on error with error                ${False}      ${True}     ${False}        ${True}
Notify on success with success            ${True}       ${False}    ${True}         ${True}
Notify on success with error              ${True}       ${False}    ${False}        ${False}
All notify prefs disabled with success    ${False}      ${False}    ${True}         ${False}

*** Keywords ***
Notify Pref Should Work
    [Arguments]    ${on_success}    ${on_error}    ${job_succeeds}
    ...            ${should_have_called_notify_program}

    # make notify program's expected output
    ${expected_output}=    Set Variable    succeeded: ${job_succeeds}, status: Good
    ${output_file}=    Make Tempfile

    # make & install jobfile
    ${cmd}=    Set Variable If    ${job_succeeds}    exit 0    exit 1
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    ...    notify_on_success=${on_success}
    ...    notify_on_error=${on_error}
    ...    notify_output_path=${output_file}
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed

    # wait
    Sleep    3s    reason=Wait for job to run

    # test
    Nothing Has Crashed
    Run Keyword If    ${should_have_called_notify_program}
    ...            File Should Have Contents    ${output_file}    ${expected_output}
    ...    ELSE    File Should Not Exist    ${output_file}
