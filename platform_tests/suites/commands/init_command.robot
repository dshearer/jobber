*** Setting ***
Library          OperatingSystem
Library          testlib.py
Resource         keywords.robot
Test Setup       Setup
Test Teardown    Teardown
Force Tags       test    cmd    init

*** Test Cases ***
Init Command
    # check initial condition
    Jobfile For Root Should Not Exist

    # do command
    Jobber Init

    # check for jobfile
    Jobfile For Root Should Exist
