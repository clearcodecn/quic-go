package self_test

import (
	"context"
	"fmt"
	"io"
	"net"

	quic "github.com/clearcodecn/quic-go"
	"github.com/clearcodecn/quic-go/internal/handshake"
	"github.com/clearcodecn/quic-go/internal/protocol"
	"github.com/clearcodecn/quic-go/logging"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sentHeaders     []*logging.ExtendedHeader
	receivedHeaders []*logging.ExtendedHeader
)

func countKeyPhases() (sent, received int) {
	lastKeyPhase := protocol.KeyPhaseOne
	for _, hdr := range sentHeaders {
		if hdr.IsLongHeader {
			continue
		}
		if hdr.KeyPhase != lastKeyPhase {
			sent++
			lastKeyPhase = hdr.KeyPhase
		}
	}
	lastKeyPhase = protocol.KeyPhaseOne
	for _, hdr := range receivedHeaders {
		if hdr.IsLongHeader {
			continue
		}
		if hdr.KeyPhase != lastKeyPhase {
			received++
			lastKeyPhase = hdr.KeyPhase
		}
	}
	return
}

type keyUpdateConnTracer struct {
	connTracer
}

func (t *keyUpdateConnTracer) SentPacket(hdr *logging.ExtendedHeader, size logging.ByteCount, ack *logging.AckFrame, frames []logging.Frame) {
	sentHeaders = append(sentHeaders, hdr)
}

func (t *keyUpdateConnTracer) ReceivedPacket(hdr *logging.ExtendedHeader, size logging.ByteCount, frames []logging.Frame) {
	receivedHeaders = append(receivedHeaders, hdr)
}

var _ = Describe("Key Update tests", func() {
	var server quic.Listener

	runServer := func() {
		var err error
		server, err = quic.ListenAddr("localhost:0", getTLSConfig(), nil)
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			sess, err := server.Accept(context.Background())
			Expect(err).ToNot(HaveOccurred())
			str, err := sess.OpenUniStream()
			Expect(err).ToNot(HaveOccurred())
			defer str.Close()
			_, err = str.Write(PRDataLong)
			Expect(err).ToNot(HaveOccurred())
		}()
	}

	It("downloads a large file", func() {
		origKeyUpdateInterval := handshake.KeyUpdateInterval
		defer func() { handshake.KeyUpdateInterval = origKeyUpdateInterval }()
		handshake.KeyUpdateInterval = 1 // update keys as frequently as possible

		runServer()
		sess, err := quic.DialAddr(
			fmt.Sprintf("localhost:%d", server.Addr().(*net.UDPAddr).Port),
			getTLSClientConfig(),
			getQuicConfig(&quic.Config{Tracer: newTracer(func() logging.ConnectionTracer { return &keyUpdateConnTracer{} })}),
		)
		Expect(err).ToNot(HaveOccurred())
		str, err := sess.AcceptUniStream(context.Background())
		Expect(err).ToNot(HaveOccurred())
		data, err := io.ReadAll(str)
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(Equal(PRDataLong))
		Expect(sess.CloseWithError(0, "")).To(Succeed())

		keyPhasesSent, keyPhasesReceived := countKeyPhases()
		fmt.Fprintf(GinkgoWriter, "Used %d key phases on outgoing and %d key phases on incoming packets.\n", keyPhasesSent, keyPhasesReceived)
		Expect(keyPhasesReceived).To(BeNumerically(">", 10))
		Expect(keyPhasesReceived).To(BeNumerically("~", keyPhasesSent, 2))
	})
})
