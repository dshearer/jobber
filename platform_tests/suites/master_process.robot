*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test

*** Test Cases ***
Kill Master Process
    # kill it
    Kill Master Proc

    # check whether there are still runner processes
    There Should Be No Runner Procs

    # restart jobber service so that Teardown doesn't fail
    Restart Service
