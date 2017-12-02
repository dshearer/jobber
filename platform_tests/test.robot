*** Setting ***
Library          OperatingSystem
Library          testlib.py
Test Setup       Setup
Test Teardown    Teardown

*** Test Cases ***
Basic
    # make jobfile for root
    ${root_expected_output}=    Set Variable    Hello
    ${root_output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n '${root_expected_output}' > ${root_output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    ${num_jobs}=    Install Root Jobfile    ${jobfile}
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load root's jobs
    
    # make jobfile for normal user
    ${normuser_expected_output}=    Set Variable    Goodbye
    ${normuser_output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n '${normuser_expected_output}' > ${normuser_output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    ${root_actual_output}=    Get File    ${root_output_file}
    Should Be Equal    ${root_expected_output}    ${root_actual_output}    msg=root's job didn't run
    ${normuser_actual_output}=    Get File    ${normuser_output_file}
    Should Be Equal    ${normuser_expected_output}    ${normuser_actual_output}    msg=Normuser's job didn't run

Privilege Separation
    # make jobfile for normal user
    ${output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n 'Hello' > ${output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    
    # change owner and mode of output file
    Chown    ${output_file}    root
    Chmod    ${output_file}    0600
    
    # install jobfile
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # give it time to run
    Sleep    3s    reason=Wait for job to run
    
    # test
    ${tmp}=    Get File    ${output_file}
    Length Should Be    ${tmp}    0    msg=Normuser was able to modify root's file

Notify On Error
    # make notify program
    ${expected_output}=    Set Variable    Hello
    ${output_file}=    Make Tempfile
    ${notify_prog}=    Make Tempfile
    Create File    ${notify_prog}    \#!/bin/sh\necho -n '${expected_output}' > ${output_file}
    Chmod    ${notify_prog}    0755
    
    # make & install jobfile
    ${jobfile}=    Make Jobfile    TestJob    exit 1    notify_prog=${notify_prog}
    Install Root Jobfile    ${jobfile}
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    ${actual_output}=    Get File    ${output_file}
    Should Be Equal    ${expected_output}    ${actual_output}

List Command
    # make jobfile for root
    ${jobfile}=    Make Jobfile    TestJob1    exit 0
    ${num_jobs}=    Install Root Jobfile    ${jobfile}
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load root's jobs
    
    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob2    exit 0
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # test 'jobber list' as root
    Jobber List as Root Should Return    TestJob1
    
    # test 'jobber list' as normuser
    Jobber List as Normuser Should Return    TestJob2
    
    # test 'jobber list -a' as root
    Jobber List as Root Should Return    TestJob1,TestJob2    all_users=True

Pause And Resume Commands
    # make & install jobfile
    ${output_file}=    Make Tempfile
    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Root Jobfile    ${jobfile}
    
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

Test Command
    # make & install jobfile
    ${output_file}=    Make Tempfile
    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Root Jobfile    ${jobfile}
    
    # pause it
    Pause Job    TestJob
    
    # It's possible that the job is running at this very moment, so
    # wait a second to let it finish.
    Sleep    1s
    
    # get latest output (if any)
    ${output_1}=    Get File    ${output_file}
    
    # call 'jobber test'
    Test Job    TestJob
    
    # check whether it ran again
    ${output_2}=    Get File    ${output_file}
    Should Not Be Equal    ${output_1}    ${output_2}    msg=Job did not run
    

*** Keyword ***
Setup
    Restart Service
    Make Tempfile Dir

Teardown
    Rm Tempfile Dir
    Rm Jobfiles