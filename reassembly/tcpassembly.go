// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package reassembly provides TCP stream re-assembly.
//
// The reassembly package implements uni-directional TCP reassembly, for use in
// packet-sniffing applications.  The caller reads packets off the wire, then
// presents them to an Assembler in the form of gopacket layers.TCP packets
// (github.com/dreadl0ck/gopacket, github.com/dreadl0ck/gopacket/layers).
//
// The Assembler uses a user-supplied
// StreamFactory to create a user-defined Stream interface, then passes packet
// data in stream order to that object.  A concurrency-safe StreamPool keeps
// track of all current Streams being reassembled, so multiple Assemblers may
// run at once to assemble packets while taking advantage of multiple cores.
//
// TODO: Add simplest example
package reassembly

import (
	"flag"
	"github.com/dreadl0ck/gopacket"
)

// TODO:
// - push to Stream on Ack
// - implement chunked (cheap) reads and Reader() interface
// - better organize file: split files: 'mem', 'misc' (seq + flow)

var defaultDebug = false

// TODO: add to flagSet
var debugLog = flag.Bool("assembly_debug_log", defaultDebug, "If true, the github.com/dreadl0ck/gopacket/reassembly library will log verbose debugging information (at least one line per packet)")

// TCPAssemblyStats provides some figures for a ScatterGather
type TCPAssemblyStats struct {
	// For this ScatterGather
	Chunks  int
	Packets int
	// For the half connection, since last call to ReassembledSG()
	QueuedBytes    int
	QueuedPackets  int
	OverlapBytes   int
	OverlapPackets int
}

// ScatterGather is used to pass reassembled data and metadata of reassembled
// packets to a Stream via ReassembledSG
type ScatterGather interface {
	// Returns the length of available bytes and saved bytes
	Lengths() (int, int)
	// Returns the bytes up to length (shall be <= available bytes)
	Fetch(length int) []byte
	// Tell to keep from offset
	KeepFrom(offset int)
	// Return CaptureInfo of packet corresponding to given offset
	CaptureInfo(offset int) gopacket.CaptureInfo
	// Return some info about the reassembled chunks
	Info() (direction TCPFlowDirection, start bool, end bool, skip int)
	// Return some stats regarding the state of the stream
	Stats() TCPAssemblyStats
}

// byteContainer is either a page or a livePacket
type byteContainer interface {
	getBytes() []byte
	length() int
	convertToPages(*pageCache, int, AssemblerContext) (*page, *page, int)
	captureInfo() gopacket.CaptureInfo
	assemblerContext() AssemblerContext
	release(*pageCache) int
	isStart() bool
	isEnd() bool
	getSeq() Sequence
	isPacket() bool
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
