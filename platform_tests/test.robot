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
    Nothing Has Crashed
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load root's jobs
    
    # make jobfile for normal user
    ${normuser_expected_output}=    Set Variable    Goodbye
    ${normuser_output_file}=    Make Tempfile
    ${cmd}=    Set Variable    echo -n '${normuser_expected_output}' > ${normuser_output_file}
    ${jobfile}=    Make Jobfile    TestJob    ${cmd}
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    Nothing Has Crashed
    ${root_actual_output}=    Get File    ${root_output_file}
    Should Be Equal    ${root_expected_output}    ${root_actual_output}    msg=root's job didn't run
    ${normuser_actual_output}=    Get File    ${normuser_output_file}
    Should Be Equal    ${normuser_expected_output}    ${normuser_actual_output}    msg=Normuser's job didn't run

Log Path Preference
    ${log_path}=    Set Variable    /home/normuser/.jobber-log
    File Should Not Exist    ${log_path}
    
    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob    exit 0
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    File Should Exist    ${log_path}    msg=Log file was not created
    File Should Not Be Empty    ${log_path}    msg=Log file is empty

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
    Nothing Has Crashed
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # give it time to run
    Sleep    3s    reason=Wait for job to run
    
    # test
    Nothing Has Crashed
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
    Nothing Has Crashed
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # test
    Nothing Has Crashed
    ${actual_output}=    Get File    ${output_file}
    Should Be Equal    ${expected_output}    ${actual_output}

List Command
    # make jobfile for root
    ${jobfile}=    Make Jobfile    TestJob1    exit 0
    ${num_jobs}=    Install Root Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load root's jobs
    
    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob2    exit 0
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    # test 'jobber list' as root
    Jobber List as Root Should Return    TestJob1
    Nothing Has Crashed
    
    # test 'jobber list' as normuser
    Jobber List as Normuser Should Return    TestJob2
    Nothing Has Crashed
    
    # test 'jobber list -a' as root
    Jobber List as Root Should Return    TestJob1,TestJob2    all_users=True
    Nothing Has Crashed

Log Command
    # check initial 'jobber log' output
    ${logs}=    Jobber Log as Root
    Nbr of Lines in String Should Be    ${logs}    1
    ${logs}=    Jobber Log as Normuser
    Nbr of Lines in String Should Be    ${logs}    1
    ${logs}=    Jobber Log as Root    all_users=${True}
    Nbr of Lines in String Should Be    ${logs}    1
    
    # make jobfile for root
    ${jobfile}=    Make Jobfile    TestJob1    exit 0
    ${num_jobs}=    Install Root Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load root's jobs
    
    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob2    exit 0
    ${num_jobs}=    Install Normuser Jobfile    ${jobfile}
    Nothing Has Crashed
    Should Be Equal as Integers    1    ${num_jobs}    msg=Failed to load normuser's jobs
    
    Sleep    3s    reason=Give jobs time to run
    
    # check 'jobber log' output
    ${logs}=    Jobber Log as Root
    Nbr of Lines in String Should Be Greater than    ${logs}    1
    ${logs}=    Jobber Log as Normuser
    Nbr of Lines in String Should Be Greater than    ${logs}    1
    ${logs}=    Jobber Log as Root    all_users=${True}
    Nbr of Lines in String Should Be Greater than    ${logs}    1

Pause And Resume Commands
    # make & install jobfile
    ${output_file}=    Make Tempfile
    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Root Jobfile    ${jobfile}
    Nothing Has Crashed
    
    # wait
    Sleep    3s    reason=Wait for job to run
    
    # pause it
    Pause Job    TestJob
    Nothing Has Crashed
    
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
    Nothing Has Crashed
    
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
    Nothing Has Crashed
    
    # pause it
    Pause Job    TestJob
    Nothing Has Crashed
    
    # It's possible that the job is running at this very moment, so
    # wait a second to let it finish.
    Sleep    1s
    
    # get latest output (if any)
    ${output_1}=    Get File    ${output_file}
    
    # call 'jobber test'
    Test Job    TestJob
    Nothing Has Crashed
    
    # check whether it ran again
    ${output_2}=    Get File    ${output_file}
    Should Not Be Equal    ${output_1}    ${output_2}    msg=Job did not run

Init Command
    # check initial condition
    Jobfile For Root Should Not Exist
    
    # do command
    Jobber Init
    
    # check for jobfile
    Jobfile For Root Should Exist

Kill Master Process
    # kill it
    Kill Master Proc
    
    # check whether there are still runner processes
    There Should Be No Runner Procs

    # restart jobber service so that Teardown doesn't fail
    Restart Service

Prefs File Excludes User
    # make prefs that exclude normal user
    Set Prefs    exclude_users=normuser
    Restart Service
    Nothing Has Crashed
    
    # so teardown doesn't fail
    ${runner_procs}=    Runner Proc Info
    Set Test Variable    ${runner_procs}
    
    # make jobfile for normal user
    ${jobfile}=    Make Jobfile    TestJob    exit 0
    Install Normuser Jobfile    ${jobfile}    reload=${False}
    
    # test
    Jobberrunner Should Not Be Running For User    normuser

Default Prefs Is Installed
    # test
    Prefs File Should Exist

Random Time Spec
    # make jobfile with random time spec
    ${time_spec}=    Set Variable    0 0 R5-8
    ${jobfile}=    Make Jobfile    TestJob    exit 0    time=${time_spec}
    ${num_jobs}=    Install Root Jobfile    ${jobfile}
    
    # test
    Nothing Has Crashed
    Should Be Equal As Integers    1    ${num_jobs}    msg=Failed to load root's jobs

*** Keyword ***
Setup
    Restart Service
    Runner Procs Should Not Have TTY
    Make Tempfile Dir
    ${runner_procs}=    Runner Proc Info
    Set Test Variable    ${runner_procs}

Teardown
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