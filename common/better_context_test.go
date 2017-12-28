package common

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func job(ctx BetterContext, jobNbr int, results chan<- int) {
	defer ctx.Finish()

	n := rand.Intn(5)
	select {
	case <-time.After(time.Duration(n) * time.Second):
		fmt.Printf("Job %v: publishing result\n", jobNbr)
		results <- n
	case <-ctx.Done():
		fmt.Printf("Job %v: cancelled\n", jobNbr)
	}
	fmt.Printf("Job %v: finishing\n", jobNbr)
}

func TestWaitForChildren(t *testing.T) {
	nbrJobs := 5

	// make main context
	ctx, _ := MakeChildContext(context.Background())
	defer ctx.Finish()

	// spawn jobs
	results := make(chan int, nbrJobs)
	for i := 0; i < nbrJobs; i++ {
		// spawn job with its own context
		jobCtx, _ := MakeChildContext(ctx)
		go job(jobCtx, i, results)
	}

	// wait for jobs to finish
	fmt.Println("Waiting for children")
	ctx.WaitForChildren()

	// check results
	var resultArr []int
	close(results)
	for r := range results {
		resultArr = append(resultArr, r)
	}
	require.Equal(t, nbrJobs, len(resultArr))
}
