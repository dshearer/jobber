*** Setting ***
Library          OperatingSystem
Library          testlib.py
Library          socketsinklib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    prefs    socketsink

*** Variables ***
${JOB_NAME}             TestJob
${TCP_SOCKET_ADDR}      :1234
${UNIX_SOCKET_PATH}     /tmp/test.sock

*** Test Cases ***
TCP Socket Sink Works
    # install jobfile with socket result sink
    ${jobfile}=    Make Jobfile    job_name=${JOB_NAME}    cmd=exit 0
    ...                            notify_on_success=${True}    tcp_result_sink_port=${TCP_SOCKET_ADDR}
    Install Jobfile    ${jobfile}    for_root=${True}

    # test
    Run Recs Should Appear on Socket    tcp    ${TCP_SOCKET_ADDR}

    # remove socket sink
    ${jobfile}=    Make Jobfile    job_name=${JOB_NAME}    cmd=exit 0
    Install Jobfile    ${jobfile}    for_root=${True}

    # test
    Jobber Procs Should Not Have Inet Sockets

Unix Socket Sink Works
    # install jobfile with socket result sink
    ${jobfile}=    Make Jobfile    job_name=${JOB_NAME}    cmd=exit 0
    ...                            notify_on_success=${True}    unix_result_sink_path=${UNIX_SOCKET_PATH}
    Install Jobfile    ${jobfile}    for_root=${True}

    # test
    Run Recs Should Appear on Socket    unix    ${UNIX_SOCKET_PATH}

    # remove socket sink
    ${jobfile}=    Make Jobfile    job_name=${JOB_NAME}    cmd=exit 0
    Install Jobfile    ${jobfile}    for_root=${True}

    # test
    File Should Not Exist    ${UNIX_SOCKET_PATH}


*** Keywords ***
Run Recs Should Appear on Socket
    [Arguments]    ${proto}    ${addr}

    # spawn process that writes data from socket to disk
    ${output_file}=    Make Tempfile
    ${proc}=    Dump Socket To Disk    ${proto}    ${addr}    ${output_file}

    # wait
    Sleep    3s

    # check data from socket
    Terminate Process    ${proc}
    File Should Exist    ${output_file}
    ${output_file_contents}=    Get File    ${output_file}
    Nbr of Lines in String Should Be Greater Than    ${output_file_contents}    2
