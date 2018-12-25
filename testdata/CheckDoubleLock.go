package check2

import (
	"fmt"
	"sync"
)

var r sync.Mutex
var rw sync.RWMutex
var r2 sync.Mutex


/* test for SA2005 */

func fn7() {
	r.Lock()
	i := 1
	fmt.Println(i)
	r.Lock()
}

func fn8() {
	rw.Lock()
	i := 1
	fmt.Println(i)
	rw.Lock()
}

func fn9() {
	r.Lock()
	i := 1
	fmt.Println(i)
	rw.Lock()
}

func fn10() {
	r.Lock()
	rw.RLock()
	i := 1
	fmt.Println(i)
	r.Unlock()
	r.Lock()
	rw.RLock()
}


func fn11() {
	r.Lock()
	i := 0
	fmt.Println(i)
	r.Lock()
}

func fn12() {
	r.Lock()
	i := 0
	fmt.Println(i)
	rw.Lock()
}

func fn133_(i int) int {
	if i > 100 {
		r.Lock()
		defer r.Unlock()
		i = i + 1
	}
	return i
}

func fn13_(i int) int {

	if i > 10 {
		fn133_(i)
	} else {
		fmt.Println('a')
	}

	return 0
}

func fn13() {
	i := 0
	r.Lock()
	defer r.Unlock()
	i = fn13_(i)
}
//
func fn14() {
	rw.RLock()
	defer rw.RUnlock()
	i := 0
	fmt.Println(i)
	rw.Lock()
	i = 0
	fmt.Println(i)
	rw.Unlock()
}
//
//
func fn15() {
	rw.RLock()
	i := 0
	fmt.Println(i)
	rw.RUnlock()
	rw.Lock()
	rw.Unlock()
}

func fn15_() {
	r.Lock()
	i := 0
	fmt.Println(i)
	r.Unlock()
}
////
func fn16(a int) {
	i := 0
	r.Lock()
	i = a
	if i >= 0 {
		r.Lock()
	}
	r.Unlock()
}
//
func fn17(a int) {
	i := 0
	r.Lock()
	i = a
	r.Unlock()
	if i >= 0 {
		r.Lock()
		i += 10
		r.Unlock()
	}
	r.Unlock()
}

// double lock in a loop
func fn18(a int) {

	i := 0

	b := 10

	ch := make(chan int)

	for {

		r.Lock()
		i += 1

		if a + b > 100 {

			fn15()

		} else {
			b += 10
		}

		if a > 10 {

			break

		} else {

			continue
		}

		fmt.Println(a)

		select {
		case ch<- 1:
			fmt.Println("write to channel")

		case b = <-ch:
			fmt.Println("read from channel")
		default:
			fmt.Println("operate channel error")
		}

		r.Unlock()
		a = 10
		fmt.Println(a)
	}
}

// double lock in a loop
func fn19(a int) {

	i := 0

	b := 10

	ch := make(chan int)

	r.Lock()
	a += 10
	defer r.Unlock()

	for j:=0 ; j < 100; j++{

		i += 1
		if a + b > 100 {
			fn15_()
			fmt.Println("write to channel")
		} else {
			b += 10
		}

		if a > 10 {
			break

		} else if b > 100 {
			continue
		}

		fmt.Println(a)

		select {
		case ch<- 1:
			fmt.Println("write to channel")

		case b = <-ch:
			fmt.Println("read from channel")
		default:
			fmt.Println("operate channel error")
		}

		r.Lock()
		a = 10
		fmt.Println(a)
		r.Unlock()
	}
}

func fn20() {
	i := 10

	b := 10

	//ch := make(chan int)

	for {

		r.Lock()

		for j:=0; j < i; j++ {
			fmt.Println(j)
			if j > 5 {
				j += 2
			}
		}

		r.Unlock()
		fmt.Println(b)
		r.Lock()

		for j:=0; j < i; j++ {
			fmt.Println(j)
			if j > 5 {
				j += 2
			}
		}
		r.Unlock()

	}
}
