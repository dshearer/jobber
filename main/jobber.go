package main

import (
    "github.com/dshearer/jobber"
    "os"
    "fmt"
)

func main() {
    // read jobs
    f, err := os.Open("go_workspace/src/github.com/dshearer/jobber/example.json")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    jobs, err := jobber.ReadJobFile(f)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    for _, job := range jobs {
        fmt.Println("Job:", job)
    }
    
    // run them
	manager := jobber.JobManager{Shell: "/bin/ksh"}
	manager.Jobs = jobs
	manager.Go()
}
