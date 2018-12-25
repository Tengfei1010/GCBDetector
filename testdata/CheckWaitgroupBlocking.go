package check1

import (
	"fmt"
	"sync"
)

/* test for SA2006 */
func fn21(array []int) {

	var wg sync.WaitGroup

	wg.Add(len(array))
	for _, a := range array {
		go func() {

			fmt.Println(a)
			wg.Done()
		}()
		wg.Wait()
	}
}
