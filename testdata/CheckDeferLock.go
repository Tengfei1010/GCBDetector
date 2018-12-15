package pkg

import (
	"sync"
	"fmt"
)

var r sync.Mutex
var rw sync.RWMutex

//func fn1() {
//	r.Lock()
//	defer r.Lock() // MATCH /deferring Lock right after having locked already; did you mean to defer Unlock/
//}
//
//func fn2() {
//	r.Lock()
//	defer r.Unlock()
//}
//
//func fn3() {
//	println("")
//	defer r.Lock()
//}
//
//func fn4() {
//	rw.RLock()
//	defer rw.RLock() // MATCH /deferring RLock right after having locked already; did you mean to defer RUnlock/
//}
//
//func fn5() {
//	rw.RLock()
//	defer rw.Lock()
//}
//
//func fn6() {
//	r.Lock()
//	defer rw.Lock()
//}


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