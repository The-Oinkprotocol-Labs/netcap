/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package decoder

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/ip4defrag"

	"github.com/dreadl0ck/netcap/reassembly"
)

var streamFactory = newStreamFactory()

func newStreamFactory() *tcpConnectionFactory {
	f := &tcpConnectionFactory{
		defragger:  ip4defrag.NewIPv4Defragmenter(),
		fsmOptions: reassembly.TCPSimpleFSMOptions{},
	}
	f.StreamPool = reassembly.NewStreamPool(f)

	return f
}

func GetStreamPool() *reassembly.StreamPool {
	return streamFactory.StreamPool
}

/*
 * The TCP factory: returns a new Connection
 */

// internal data structure to handle new network streams
// and spawn the stream decoder routines for processing the data
type tcpConnectionFactory struct {
	wg            sync.WaitGroup
	decodeHTTP    bool
	decodePOP3    bool
	decodeSSH     bool
	numActive     int64
	streamReaders []streamReader

	defragger  *ip4defrag.IPv4Defragmenter
	StreamPool *reassembly.StreamPool
	fsmOptions reassembly.TCPSimpleFSMOptions

	sync.Mutex
}

// New handles a new stream received from the assembler
// this is the entry point for new network streams
// depending on the used ports, a dedicated stream reader instance will be started and subsequently fed with new data from the stream.
func (factory *tcpConnectionFactory) New(net, transport gopacket.Flow, ac reassembly.AssemblerContext) reassembly.Stream {
	logReassemblyDebug("* NEW: %s %s\n", net, transport)

	stream := &tcpConnection{
		net:         net,
		transport:   transport,
		tcpstate:    reassembly.NewTCPSimpleFSM(factory.fsmOptions),
		ident:       filepath.Clean(fmt.Sprintf("%s-%s", net, transport)),
		optchecker:  reassembly.NewTCPOptionCheck(),
		firstPacket: ac.GetCaptureInfo().Timestamp,
	}

	// do not write encrypted HTTP streams to disk for now
	// if stream.isHTTPS {
	//	 return stream
	// }

	stream.decoder = &tcpReader{
		parent: stream,
	}
	stream.client = stream.newTCPStreamReader(true)
	stream.server = stream.newTCPStreamReader(false)

	factory.wg.Add(2)

	factory.Lock()
	factory.streamReaders = append(factory.streamReaders, stream.client)
	factory.streamReaders = append(factory.streamReaders, stream.client)
	factory.numActive += 2
	factory.Unlock()

	// launch stream readers
	go stream.client.Run(factory)
	go stream.server.Run(factory)

	return stream
}

// waitGoRoutines waits until the goroutines launched to process TCP streams are done
// this will block forever if there are streams that are never shutdown (via RST or FIN flags)
func (factory *tcpConnectionFactory) waitGoRoutines() {
	if !c.Quiet {
		factory.Lock()
		fmt.Println("\nwaiting for", factory.numActive, "flows")
		factory.Unlock()
	}

	factory.wg.Wait()
}

// context is the assembler context
type context struct {
	CaptureInfo gopacket.CaptureInfo
}

// GetCaptureInfo returns the gopacket.CaptureInfo from the context
func (c *context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}
