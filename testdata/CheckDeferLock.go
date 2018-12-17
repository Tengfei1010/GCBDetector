package pkg

import (
	"fmt"
	"sync"
)

var r sync.Mutex
var rw sync.RWMutex
var r2 sync.Mutex

func fn1() {
	r.Lock()
	defer r.Lock() // MATCH /deferring Lock right after having locked already; did you mean to defer Unlock/
}

func fn2() {
	r.Lock()
	defer r.Unlock()
}

func fn3() {
	println("")
	defer r.Lock()
}

func fn4() {
	rw.RLock()
	defer rw.RLock() // MATCH /deferring RLock right after having locked already; did you mean to defer RUnlock/
}

func fn5() {
	rw.RLock()
	defer rw.Lock()
}

func fn6() {
	r.Lock()
	defer rw.Lock()
}


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
//
///* **************** */
//
///* test for SA2005 */
//
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

func fn13_(i int) int {
	r.Lock()
	defer r.Unlock()
	i = i + 1
	return i
}

func fn13() {
	i := 0
	r.Lock()
	defer r.Unlock()
	i = fn13_(i)
}

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

func fn15() {
	rw.RLock()
	i := 0
	fmt.Println(i)
	rw.RUnlock()
	rw.Lock()
}

func fn16(a int) {
	i := 0
	r.Lock()
	i = a
	if i >= 0 {
		r.Lock()
	}
	r.Unlock()
}

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