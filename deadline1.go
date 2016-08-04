package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	t := time.Now()
	t = t.Add(3 * time.Second)
	ctx, _ := context.WithDeadline(context.Background(), t)
	go doStuff(ctx)
	time.Sleep(5 * time.Second)
}

func doStuff(ctx context.Context) {
	// ...Doing some work
	deadline, ok := ctx.Deadline()
	if ok {
		fmt.Println("DEADLINE:", deadline)
		fmt.Println("NOW:", time.Now())

		diff := deadline.Sub(time.Now())
		fmt.Println("diff:", diff)
		if diff > 0 {
			fmt.Println("Not enough time left in deadline")
		}
	}

	fmt.Println("Do work...")
}
