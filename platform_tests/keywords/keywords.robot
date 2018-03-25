*** Keyword ***
Setup
    Restart Service
    Runner Procs Should Not Have TTY
    Jobber Procs Should Not Have Inet Sockets
    Make Tempfile Dir
    ${runner_procs}=    Runner Proc Info
    Set Test Variable    ${runner_procs}

Teardown
    Runner Procs Should Not Have TTY
    Jobber Procs Should Not Have Inet Sockets
    Rm Tempfile Dir
    Rm Jobfiles
    Run Keyword If Test Failed    Print Debug Info
    Remove Files    /root/.jobber-log    /home/normuser/.jobber-log
    Nbr Of Runner Procs Should Be Same    ${runner_procs}
    Restore Prefs

Nothing Has Crashed
    jobbermaster Has Not Crashed
    jobberrunner for Root Has Not Crashed
    jobberrunner for Normuser Has Not Crashed

File Should Have Contents
    [Arguments]    ${path}    ${contents}    ${msg}=${None}
    File Should Exist    ${path}    msg=${msg}
    ${actual_contents}=    Get File    ${path}
    Should Be Equal    ${contents}    ${actual_contents}    msg=${msg}
