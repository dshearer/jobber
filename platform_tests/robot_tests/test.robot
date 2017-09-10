*** Setting ***
Library          OperatingSystem
Library          testlib.py
Test Setup       Setup
Test Teardown    Teardown

*** Test Cases ***
Basic
    # make jobfile
    ${expected_output}=    Set Variable    Hello
    ${output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n '${expected_output}' > ${output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    
    # install jobfile
    Install Jobfile    ${jobfile}
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    ${actual_output}=    Get File    ${output_file}
    Should Be Equal    ${expected_output}    ${actual_output}

Notify On Error
    # make notify program
    ${expected_output}=    Set Variable    Hello
    ${output_file}=    Make Tempfile
    ${notify_prog}=    Make Tempfile
    Create File    ${notify_prog}    \#!/bin/sh\necho -n '${expected_output}' > ${output_file}
    Chmod    ${notify_prog}    0755
    
    # make & install jobfile
    ${jobfile}=    Make Jobfile    TestJob    exit 1    notify_prog=${notify_prog}
    Install Jobfile    ${jobfile}
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    ${actual_output}=    Get File    ${output_file}
    Should Be Equal    ${expected_output}    ${actual_output}

Pause And Resume
    # make & install jobfile
    ${output_file}=    Make Tempfile
    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Jobfile    ${jobfile}
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # pause it
    Pause Job    TestJob
    
    # It's possible that the job is running at this very moment, so
    # wait a second to let it finish.
    Sleep    1s
    
    # get latest output
    ${output_1}=    Get File    ${output_file}
    
    # wait
    Sleep    5s    reason=See if job will run again
    
    # check whether it ran again
    ${output_2}=    Get File    ${output_file}
    Should Be Equal    ${output_1}    ${output_2}    msg=Job ran when paused
    
    # resume it
    Resume Job    TestJob
    
    # wait
    Sleep    5s    reason=See if job will run again
    
    # check whether it ran again
    ${output_3}=    Get File    ${output_file}
    Should Not Be Equal    ${output_1}    ${output_3}    msg=Job did not run when resumed

*** Keyword ***
Setup
    Make Tempfile Dir

Teardown
    Rm Tempfile Dir
    Rm Jobfile