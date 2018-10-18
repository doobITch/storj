// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package kademlia

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"storj.io/storj/pkg/node"
	"storj.io/storj/pkg/pb"
)

//IDFromBinStr turns a string like '110001' into a string like 'a'
func BinStr(s string) string {
	b := []byte(strings.Repeat("0", 8-len(s)%8) + s)
	a := make([]byte, len(b)/8)
	for i := 0; i < len(b); i++ {
		a[i/8] |= ((b[i] - '0') << uint(7-i%8))
	}
	return string(a)
}


func TestXorQueue(t *testing.T) {
	target := node.IDFromString("a") //0001 (shortened from 1100001)
	testValues := []string {"c", "f", "g", "h"} //0011, 0110, 0111, 1000
	expectedPriority := []int{2, 6, 7, 9} // 0010=>2, 0111=>7, 0110=>6, 1001=>9
	expectedIds := []string{"c", "g", "f", "h"}

	nodes := make([]*pb.Node, len(testValues))
	for i, value := range testValues {
		nodes[i] = &pb.Node{ Id: value }
	}
	pq := NewXorQueue(nodes, target)
	
	for i := 0; pq.Len() > 0; i++ {
		node, priority := pq.PopClosest()
		assert.Equal(t, big.NewInt(int64(expectedPriority[i])), priority)
		assert.Equal(t, expectedIds[i], node.Id)
	}
}
func TestXorQueue(t *testing.T) {
	target := node.ID(BinStr("0001"))
	testValues := []string {"0011", "0110", "0111", "1000"} //0011, 0110, 0111, 1000
	expectedPriority := []int{2, 6, 7, 9} // 0010=>2, 0111=>7, 0110=>6, 1001=>9
	expectedIds := []string{"0011", "0111", "0110", "1000"}

	nodes := make([]*pb.Node, len(testValues))
	for i, value := range testValues {
		nodes[i] = &pb.Node{ Id: BinStr(value) }
	}
	pq := NewXorQueue(nodes, &target)
	
	for i := 0; pq.Len() > 0; i++ {
		node, priority := pq.PopClosest()
		assert.Equal(t, big.NewInt(int64(expectedPriority[i])), priority)
		assert.Equal(t, BinStr(expectedIds[i]), node.Id)
	}
}