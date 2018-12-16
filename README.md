## Go Concurrency Bug Detector
Go Concurrency Bug Detector Based on Golang SSA

### How to install GCBDetector
$ go get Tengfei1010/GCBDetector/...

### How to use it
$ cd $GOROOT/bin
$ staticcheck google.golang.org/grpc

### How to write your checker
Please put your checker in staticcheck/lint.go(from line 53)

-------
#### The framework is copyed from https://github.com/dominikh/go-tools, you can see the detail in https://staticcheck.io/
