# Simple promise API written in go

This is more a learning project than anything else.


# Usage

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	promise "github.com/iwyg/gopromise"
	"time"
)

func main() {
	log.Println("promise 1")

	p1 := promise.New(func(res func(interface{}), errFunc func(e error)) {
		resp, err := http.Get("https://github.com/iwyg")

		if err != nil {
			errFunc(err)
			return
		}

		res(resp)
	})

	p1.Then(func(val interface{}) {
		response := val.(*http.Response)

		defer response.Body.Close()
		bytes, err := httputil.DumpResponse(response, false)

		if err != nil {
			panic(nil)
		}
		log.Println("p1: dumping Response received from github repository…")
		log.Printf("%q", bytes)
	})

	p1.Fail(func(e error) {
		log.Println("error fetching url")
	})

	log.Println("promise 2")

	p2 := promise.New(func(res func(interface{}), errFunc func(e error)) {
		defer res(2)
		time.Sleep(time.Second * 100)
	})

	p2.WhenCancelled(func() {
		log.Println("p2 got cancelled")
	})

	p2.Then(func(val interface{}) {
		value := val.(int)
		fmt.Println("foo")
		log.Printf("value %d recieved", value)
	})

	go func() {
		defer p2.Cancel()
		time.Sleep(time.Second * 1)
	}()

	p3 := promise.New(func(res func(interface{}), errFunc func(e error)) {
		time.Sleep(time.Second * 100)
		res(3)
	})

	p3.Until(func(val interface{}) {
		value := val.(int)
		log.Printf("p3 received %d", value)
	}, time.Millisecond*100)

	p3.WhenCancelled(func() {
		log.Println("p3 got cancelled")
	})

	log.Println("promise 3")

	time.Sleep(time.Second * 5)
}


```