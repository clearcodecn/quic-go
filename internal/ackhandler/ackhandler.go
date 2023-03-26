package ackhandler

import (
	"github.com/clearcodecn/quic-go/internal/protocol"
	"github.com/clearcodecn/quic-go/internal/utils"
	"github.com/clearcodecn/quic-go/logging"
)

// NewAckHandler creates a new SentPacketHandler and a new ReceivedPacketHandler
func NewAckHandler(
	initialPacketNumber protocol.PacketNumber,
	initialMaxDatagramSize protocol.ByteCount,
	rttStats *utils.RTTStats,
	pers protocol.Perspective,
	tracer logging.ConnectionTracer,
	logger utils.Logger,
	version protocol.VersionNumber,
) (SentPacketHandler, ReceivedPacketHandler) {
	sph := newSentPacketHandler(initialPacketNumber, initialMaxDatagramSize, rttStats, pers, tracer, logger)
	return sph, newReceivedPacketHandler(sph, rttStats, logger, version)
}
