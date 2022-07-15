package singleflight

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do("test", func() (interface{}, error) {
		return "test", nil
	})

	if got, want := fmt.Sprintf("%s(%T)", v, v), "test(string)"; got != want {
		t.Errorf(fmt.Sprintf("got not equals to want. got: %s", got))
	}
	if err != nil {
		t.Errorf("error is not nil")
	}
}

func TestDoOnce(t *testing.T) {
	var g Group
	var calls int32
	var wg sync.WaitGroup
	ch := make(chan string)
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return <-ch, nil
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("test", fn)
			if err != nil {
				t.Errorf("err:%v", err)
			}

			if v.(string) != "foo" {
				t.Errorf("got: %v, want: foo", v)
			}
			wg.Done()
		}()
	}

	// block the goroutines above
	time.Sleep(100 * time.Millisecond)

	ch <- "foo"
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("got: %d, want: 1", got)
	}
}
