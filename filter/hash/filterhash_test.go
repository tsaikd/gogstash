package hash

import (
	"errors"
	"fmt"
	"github.com/tsaikd/gogstash/config/logevent"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestInitHash checks that
// 1. all hashers can be initialized
// 2. that we get error with on unknown hash
func TestInitHash(t *testing.T) {
	// check list
	for _, v := range hashAlgos {
		f := FilterConfig{}
		f.Kind = v.Name
		err := initHashConfig(&f)
		if err != nil {
			t.Errorf("Hash %s failed init: %v", v.Name, err)
		}
		if f.hash == nil {
			t.Errorf("Hash %s failed after init", v.Name)
		}
		if (f.hash == nil && f.hash32 == nil) ||
			(f.hash != nil && f.hash32 != nil) {
			t.Errorf("FilterConfig for %s: hash and hash32 are incorrect - either both nil or non-nil", v.Name)
		}
	}
	// try unknown
	f := FilterConfig{}
	f.Kind = "my-unknown-hash"
	err := initHashConfig(&f)
	if err == nil {
		t.Errorf("Unknown hash should have failed init")
	}

}

// TestHashDuplicateNames check for duplicate names
func TestHashDuplicateNames(t *testing.T) {
	list := make(map[string]bool)
	dupes := []string{}
	// go trough list
	for _, v := range getAllHashes() {
		if _, inList := list[v]; !inList {
			list[v] = false
		} else {
			if list[v] == false {
				dupes = append(dupes, v)
				list[v] = true
			}
		}
	}
	// done
	if len(dupes) > 0 {
		t.Error("Duplicate hash algo names", dupes)
	}
}

// getAllHashes returns a list of all hashes supported, used by other tests
func getAllHashes() []string {
	result := []string{}
	for _, v := range hashAlgos {
		result = append(result, v.Name)
	}
	for _, v := range hash32Algos {
		result = append(result, v.Name)
	}
	return result
}

// TestFilterConfig_makeHash runs a test on all hashes to verify that it works and that it produces the same output every time.
func TestFilterConfig_makeHash(t *testing.T) {
	const numRuns = 10000
	var wg sync.WaitGroup
	const myStringToHash = "THIS IS A TEST MESSAGE, HELLO EARTH!"
	// A list of the string above pre-hashed so we can verify that correct output is produced over time.
	// If a hash is not in the list it will not generate an error.
	knownHashes := map[string]string{
		"sha1":    "5448495320495320412054455354204d4553534147452c2048454c4c4f20454152544821da39a3ee5e6b4b0d3255bfef95601890afd80709",
		"sha256":  "5448495320495320412054455354204d4553534147452c2048454c4c4f20454152544821e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"md5":     "5448495320495320412054455354204d4553534147452c2048454c4c4f20454152544821d41d8cd98f00b204e9800998ecf8427e",
		"fnv128a": "5448495320495320412054455354204d4553534147452c2048454c4c4f204541525448216c62272e07bb014262b821756295c58d",
		"adler32": "5009eaad",
		"fnv32a":  "cc7539cd",
	}

	var totalRuns uint64
	// the test function
	myTest := func(f FilterConfig) error {
		baseResult := f.makeHash(myStringToHash) // baseline
		if known, ok := knownHashes[f.Kind]; ok {
			if known != baseResult {
				t.Errorf("%s hash does not match known result (got %s, want %s)", f.Kind, baseResult, known)
			}
		} else {
			fmt.Println("New hash", f.Kind, baseResult) // printed out so it can be added above
		}
		var myWg sync.WaitGroup
		var errCounter uint64
		var x uint64
		for x = 0; x < numRuns; x++ {
			myWg.Add(1)
			go func() {
				defer myWg.Done()
				result := f.makeHash(myStringToHash)
				if result != baseResult {
					atomic.AddUint64(&errCounter, 1)
				}
			}()
		}
		atomic.AddUint64(&totalRuns, x)
		myWg.Wait()
		if errCounter > 0 {
			msg := fmt.Sprintf("Hash %s failed %v times", f.Kind, errCounter)
			return errors.New(msg)
		}
		return nil
	} // myTest()
	// initHash prepares to call myTest
	initHash := func(hash string) {
		defer wg.Done()
		f := FilterConfig{Kind: hash}
		err := initHashConfig(&f)
		if err != nil {
			t.Errorf("Failed to init %s", hash)
		}
		err = myTest(f)
		if err != nil {
			t.Errorf("%s error %v", hash, err)
		}
	} // initHash
	// start all hashes
	for _, v := range getAllHashes() {
		wg.Add(1)
		go initHash(v)
	}
	// wait and check if we have right amount of runs
	wg.Wait()
	if (uint64(len(hashAlgos))+uint64(len(hash32Algos)))*numRuns != totalRuns {
		t.Errorf("Wrong number of test runs")
	}
}

// generateTestEvent returns an event used for testing.
// Changing anything here can lead to a test failing as its content is verified in other places.
func generateTestEvent() logevent.LogEvent {
	e := logevent.LogEvent{
		Timestamp: time.Time{},
		Message:   "THIS IS A TEST MESSAGE, HELLO EARTH!",
		Tags:      []string{"tag1", "tag2"},
		Extra: map[string]interface{}{
			"FIELD1": "text string for field 1",
			"FIELD2": 100,
			"FIELD3": 9.1,
			"FIELD4": true,
			"FIELD5": nil,
		},
	}
	return e
}

// TestGetUnhashedString tests the getUnhashedString method
func TestGetUnhashedString(t *testing.T) {
	f := FilterConfig{}
	e := generateTestEvent()
	// check with several fields
	f.Source = []string{"FIELD1", "FIELD2", "FIELD3", "FIELD4"}
	output := f.getUnhashedString(&e)
	if output != "text string for field 11009.1true" { // manually created expected result from generateTestEvent
		t.Errorf("getUnhashedString returned wrong string to hash, got %s", output)
	}
	// check with only the message field
	f.Source = []string{"message"}
	output = f.getUnhashedString(&e)
	if output != e.Message {
		t.Errorf("getUnhashedString returned wrong message, got %s", output)
	}
}

func Test_i32tob(t *testing.T) {
	type args struct {
		val uint32
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Value 0", args{0}, []byte{0, 0, 0, 0}},
		{"Value 65535", args{65535}, []byte{255, 255, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := i32tob(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("i32tob() = %v, want %v", got, tt.want)
			}
		})
	}
}
