package goroutines

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

func Run(max_workers int) {

	jobs := make(chan float64, max_workers)
	var wg sync.WaitGroup
	worker_id := 0

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		jobs <- input
		if worker_id < max_workers {
			wg.Add(1)
			worker_id++
			go func(worker_id int, jobs <-chan float64) {

				fmt.Printf("worker:%d spawning\n", worker_id)
				for job := range jobs {
					fmt.Printf("worker:%d sleep:%.1f\n", worker_id, job)
					time.Sleep(time.Duration(1000000000 * job))
				}
				fmt.Printf("worker:%d stopping\n", worker_id)
				wg.Done()

			}(worker_id, jobs)
		}
	}
	close(jobs)
	wg.Wait()
}
