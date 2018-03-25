*** Setting ***
Library          OperatingSystem
Library          testlib.py
Library          ipclib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    ipc

*** Variables ***
${OLD_JOB}=    OldJob

*** Test Cases ***
Set Job Command (Without Jobfile)
    Do Set Job Cmd    IpcJob    echo hi
    Jobber List Should Return    IpcJob    as_root=${True}
    Nothing Has Crashed

Set Job Command (With Jobfile)
    Make and Install Jobfile

    Do Set Job Cmd    IpcJob    echo hi
    Jobber List Should Return    ${OLD_JOB},IpcJob    as_root=${True}
    Nothing Has Crashed

Delete Job Command (With Jobfile)
    Make and Install Jobfile

    Do Delete Job Cmd    ${OLD_JOB}
    Jobber List Should Return    ${EMPTY}    as_root=${True}
    Nothing Has Crashed

*** Keywords ***
Make and Install Jobfile
    ${jobfile}=    Make Jobfile    ${OLD_JOB}    exit 0
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed
