// Package numtree implements radix tree data structure for 32 and 64-bit unsigned integets. Copy-on-write is used for any tree modification so old root doesn't see any change happens with tree.
package numtree

var clzTable = []uint8{4, 3, 2, 2, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0}
