package gls

import (
	"fmt"
	"sync"
	"testing"
)

func TestContexts(t *testing.T) {
	mgr1 := NewContextManager(Option{})
	mgr2 := NewContextManager(Option{})

	CheckVal := func(mgr *ContextManager, key, exp_val string) {
		val, ok := mgr.GetValue(key)
		if len(exp_val) == 0 {
			if ok {
				t.Fatalf("expected no value for key %s, got %s", key, val)
			}
			return
		}
		if !ok {
			t.Fatalf("expected value %s for key %s, got no value",
				exp_val, key)
		}
		if exp_val != val {
			t.Fatalf("expected value %s for key %s, got %s", exp_val, key,
				val)
		}

	}

	Check := func(exp_m1v1, exp_m1v2, exp_m2v1, exp_m2v2 string) {
		CheckVal(mgr1, "key1", exp_m1v1)
		CheckVal(mgr1, "key2", exp_m1v2)
		CheckVal(mgr2, "key1", exp_m2v1)
		CheckVal(mgr2, "key2", exp_m2v2)
	}

	Check("", "", "", "")
	mgr2.SetValues(Values{"key1": "val1c"}, func() {
		Check("", "", "val1c", "")
		mgr1.SetValues(Values{"key1": "val1a"}, func() {
			Check("val1a", "", "val1c", "")
			mgr1.SetValues(Values{"key2": "val1b"}, func() {
				Check("val1a", "val1b", "val1c", "")
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					Check("", "", "", "")
				}()
				Go(func() {
					defer wg.Done()
					Check("val1a", "val1b", "val1c", "")
				})
				wg.Wait()
				Check("val1a", "val1b", "val1c", "")
			})
			Check("val1a", "", "val1c", "")
		})
		Check("", "", "val1c", "")
	})
	Check("", "", "", "")
}

func ExampleContextManager_SetValues() {
	var (
		mgr            = NewContextManager(Option{})
		request_id_key = GenSym()
	)

	MyLog := func() {
		if request_id, ok := mgr.GetValue(request_id_key); ok {
			fmt.Println("My request id is:", request_id)
		} else {
			fmt.Println("No request id found")
		}
	}

	mgr.SetValues(Values{request_id_key: "12345"}, func() {
		MyLog()
	})
	MyLog()

	// Output: My request id is: 12345
	// No request id found
}

func ExampleGo() {
	var (
		mgr            = NewContextManager(Option{})
		request_id_key = GenSym()
	)

	MyLog := func() {
		if request_id, ok := mgr.GetValue(request_id_key); ok {
			fmt.Println("My request id is:", request_id)
		} else {
			fmt.Println("No request id found")
		}
	}

	mgr.SetValues(Values{request_id_key: "12345"}, func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			MyLog()
		}()
		wg.Wait()
		wg.Add(1)
		Go(func() {
			defer wg.Done()
			MyLog()
		})
		wg.Wait()
	})

	// Output: No request id found
	// My request id is: 12345
}

func BenchmarkGetValue(b *testing.B) {
	mgr := NewContextManager(Option{})
	wg := sync.WaitGroup{}
	mgr.SetValues(Values{"test_key": "test_val"}, func() {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			Go(func() {
				defer wg.Done()
				val, ok := mgr.GetValue("test_key")
				if !ok || val != "test_val" {
					b.FailNow()
				}
			})
		}
		wg.Wait()
	})
}

func BenchmarkSetValues(b *testing.B) {
	mgr := NewContextManager(Option{})
	wg := sync.WaitGroup{}
	for i := 0; i < b.N/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.SetValues(Values{"test_key": "test_val"}, func() {
				mgr.SetValues(Values{"test_key2": "test_val2"}, func() {})
			})
		}()
	}
	wg.Wait()
}

func TestExtend(t *testing.T) {
	lenCheck := func(values []Values, expected int) {
		if len(values) != expected {
			t.Fatalf("expected length %d for values length %d, got no value", expected, len(values))
		}
	}
	mgr := NewContextManager(Option{})
	lenCheck(mgr.values, initialMaxGoroutineCount)
	mgr.extend(0)
	lenCheck(mgr.values, initialMaxGoroutineCount)
	mgr.extend(initialMaxGoroutineCount)
	lenCheck(mgr.values, initialMaxGoroutineCount+extendUnit)
	mgr.extend(initialMaxGoroutineCount + extendUnit*10)
	lenCheck(mgr.values, initialMaxGoroutineCount+extendUnit*11)
}
