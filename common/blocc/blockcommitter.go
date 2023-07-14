package blocc

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/msp"
)

var blockCommitterLogger = flogging.MustGetLogger("blockCommitterLogger")

// GossipBlockCommitter is an interface that allows the ledger to notify the gossip layer
// that a block has been committed. This is used to trigger the gossip layer to broadcast
// the block to peers.
type GossipBlockCommitter interface {
	// OnBlockCommitted Gossips block after being committed
	OnBlockCommitted(txID string)
}

type GossipBlockCommitterImpl struct {
	Approver msp.SigningIdentity
}

func (g *GossipBlockCommitterImpl) OnBlockCommitted(txID string) {
	blockCommitterLogger.Debug("BLOCC: now inside blockCommitted in gossip_service.go")
	//g.GossipSvc.OnBlockCommitted(nil)
}
