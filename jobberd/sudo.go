package main

import (
    "io/ioutil"
    "os/exec"
)

type SudoResult struct {
    Stdout     []byte
    Stderr     []byte
    Succeeded  bool
}

func sudo(user string, cmdStr string, shell string, input *[]byte) (*SudoResult, *JobberError) {
    var cmd *exec.Cmd = sudo_cmd(user, cmdStr, shell);
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, &JobberError{"Failed to get pipe to stdout.", err}
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, &JobberError{"Failed to get pipe to stderr.", err}
    }
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return nil, &JobberError{"Failed to get pipe to stdin.", err}
    }
    
    // start cmd
    if err := cmd.Start(); err != nil {
        return nil, &JobberError{"Failed to execute command \"" + cmdStr + "\".", err}
    }
    
    if input != nil {
        // write input
        stdin.Write(*input)
    }
    stdin.Close()
    
    // read output
    stdoutBytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        return nil, &JobberError{"Failed to read stdout.", err}
    }
    stderrBytes, err := ioutil.ReadAll(stderr)
    if err != nil {
        return nil, &JobberError{"Failed to read stderr.", err}
    }
    
    // finish execution
    err = cmd.Wait()
    if err != nil {
        ErrLogger.Printf("sudo: %v", err)
        _, flag := err.(*exec.ExitError)
        if !flag {
            return nil, &JobberError{"Failed to execute command \"" + cmdStr + "\".", err}
        }
    }
    
    // return result
    res := &SudoResult{}
    res.Stdout = stdoutBytes
    res.Stderr = stderrBytes
    res.Succeeded = (err == nil)
    return res, nil
}
