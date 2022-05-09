package geecache

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	ConsistentRing := NewConsistent(3, func(b []byte) uint32 {
		i, _ := strconv.Atoi(string(b))
		return uint32(i)
	})

	ConsistentRing.AddNodes("2", "4", "6")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, node := range testCases {
		if tmp := ConsistentRing.Get(k); tmp != node {
			t.Errorf("Got wrong node %s, expect %s", tmp, node)
		}
	}

	ConsistentRing.AddNode("8")

	testCases["27"] = "8"

	for k, node := range testCases {
		if tmp := ConsistentRing.Get(k); tmp != node {
			t.Errorf("Got wrong node %s, expect %s", tmp, node)
		}
	}
}
