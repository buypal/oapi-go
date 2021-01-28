package logging

import "fmt"

type Printer interface {
	Printf(string, ...interface{})
}

type voidLog struct{}

func (voidLog) Printf(string, ...interface{}) {}

func Void() Printer {
	return &voidLog{}
}

func LogFunc(log Printer) func(string, ...interface{}) {
	if log != nil {
		return log.Printf
	}
	return func(a string, args ...interface{}) { fmt.Printf(a, args...) }

}
