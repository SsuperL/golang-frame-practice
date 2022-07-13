package ccache

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	hash := NewMap(3, func(key []byte) uint32 {
		v, _ := strconv.Atoi(string(key))
		return uint32(v)
	})

	hash.Add("1", "3", "5")
	testCases := map[string]string{
		"2":  "3",
		"22": "3",
		"24": "5",
		"16": "1",
		"26": "1",
	}

	for k, v := range testCases {
		a := hash.Get(k)
		if a != v {
			t.Errorf("got %v, want %v", hash.Get(k), v)
		}
	}

	// add new node
	hash.Add("6")
	testCases["16"] = "6"
	testCases["26"] = "6"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("got %v, want %v", hash.Get(k), v)
		}
	}

}
