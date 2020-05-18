*** Keyword ***
Setup
    Restart Service
    Runner Procs Should Not Have TTY
    Jobber Procs Should Not Have Inet Sockets
    Make Tempfile Dir
    ${runner_procs}=    Runner Proc Info
    Set Test Variable    ${runner_procs}

Teardown
    [Arguments]    ${check_for_socket_in_teardown}=${True}
    Runner Procs Should Not Have TTY
    Run Keyword If    ${check_for_socket_in_teardown}
    ...               Jobber Procs Should Not Have Inet Sockets
    Nbr Of Runner Procs Should Be Same    ${runner_procs}
    Rm Tempfile Dir
    Rm Jobfiles
    Run Keyword If Test Failed    Print Debug Info
    Remove Files    /root/.jobber-log    /home/normuser/.jobber-log
    Restore Prefs
    Stop Service

Nothing Has Crashed
    jobbermaster Has Not Crashed
    jobberrunner for Root Has Not Crashed
    jobberrunner for Normuser Has Not Crashed

File Should Have Contents
    [Arguments]    ${path}    ${contents}    ${msg}=${None}    ${strip_space}=${False}
    File Should Exist    ${path}    msg=${msg}
    ${actual_contents}=    Get File    ${path}
    Run Keyword If    not ${strip_space}
    ...    Should Be Equal    ${contents}    ${actual_contents}    msg=${msg}
    Run Keyword If    ${strip_space}
    ...    Should Be Equal    ${contents.strip()}    ${actual_contents.strip()}    msg=${msg}
