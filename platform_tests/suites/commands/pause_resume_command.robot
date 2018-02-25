*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test

*** Test Cases ***
Pause and Resume Commands
  # make & install jobfile
  ${output_file}=    Make Tempfile
  Make and Install Jobfile    ${output_file}

  Sleep    3s    reason=Wait for job to run

  # pause it
  Pause Job    TestJob
  Nothing Has Crashed

  # It's possible that the job is running at this very moment, so
  # wait a second to let it finish.
  Sleep    1s

  # get latest output
  ${output_1}=    Get File    ${output_file}

  Sleep    5s    reason=See if job will run again

  # check whether it ran again
  ${output_2}=    Get File    ${output_file}
  Should Be Equal    ${output_1}    ${output_2}    msg=Job ran when paused

  # resume it
  Resume Job    TestJob
  Nothing Has Crashed

  Sleep    5s    reason=See if job will run again

  # check whether it ran again
  ${output_3}=    Get File    ${output_file}
  Should Not Be Equal    ${output_1}    ${output_3}    msg=Job did not run when resumed

*** Keywords ***
Make and Install Jobfile
    [Arguments]    ${output_file}

    ${jobfile}=    Make Jobfile    TestJob    date > ${output_file}
    Install Jobfile    ${jobfile}    for_root=${True}
    Nothing Has Crashed
