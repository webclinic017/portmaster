package network

import (
	"time"

	"github.com/safing/portmaster/intel"
	"github.com/safing/portmaster/network/netutils"
	"github.com/safing/portmaster/network/packet"
	"github.com/safing/portmaster/process"
)

// Static reasons
const (
	ReasonUnknownProcess = "unknown connection owner: process could not be found"
)

// GetUnknownCommunication returns the connection to a packet of unknown owner.
func GetUnknownCommunication(pkt packet.Packet) (*Communication, error) {
	if pkt.IsInbound() {
		switch netutils.ClassifyIP(pkt.Info().Src) {
		case netutils.HostLocal:
			return getOrCreateUnknownCommunication(pkt, IncomingHost)
		case netutils.LinkLocal, netutils.SiteLocal, netutils.LocalMulticast:
			return getOrCreateUnknownCommunication(pkt, IncomingLAN)
		case netutils.Global, netutils.GlobalMulticast:
			return getOrCreateUnknownCommunication(pkt, IncomingInternet)
		case netutils.Invalid:
			return getOrCreateUnknownCommunication(pkt, IncomingInvalid)
		}
	}

	switch netutils.ClassifyIP(pkt.Info().Dst) {
	case netutils.HostLocal:
		return getOrCreateUnknownCommunication(pkt, PeerHost)
	case netutils.LinkLocal, netutils.SiteLocal, netutils.LocalMulticast:
		return getOrCreateUnknownCommunication(pkt, PeerLAN)
	case netutils.Global, netutils.GlobalMulticast:
		return getOrCreateUnknownCommunication(pkt, PeerInternet)
	case netutils.Invalid:
		return getOrCreateUnknownCommunication(pkt, PeerInvalid)
	}

	// this should never happen
	return getOrCreateUnknownCommunication(pkt, PeerInvalid)
}

func getOrCreateUnknownCommunication(pkt packet.Packet, connScope string) (*Communication, error) {
	connection, ok := GetCommunication(process.UnknownProcess.Pid, connScope)
	if !ok {
		connection = &Communication{
			Scope:                connScope,
			Entity:               (&intel.Entity{}).Init(),
			Direction:            pkt.IsInbound(),
			Verdict:              VerdictDrop,
			Reason:               ReasonUnknownProcess,
			process:              process.UnknownProcess,
			Inspect:              false,
			FirstLinkEstablished: time.Now().Unix(),
		}
		if pkt.IsOutbound() {
			connection.Verdict = VerdictBlock
		}
	}
	connection.process.AddCommunication()
	return connection, nil
}
