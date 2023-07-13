package chaincode

import (
	"crypto/tls"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/internal/peer/common"
	"github.com/hyperledger/fabric/internal/pkg/identity"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ClientConnections holds the clients for connecting to the various
// endpoints in a Fabric network.
type ClientConnections struct {
	BroadcastClient common.BroadcastClient
	DeliverClients  []pb.DeliverClient
	EndorserClients []pb.EndorserClient
	Certificate     tls.Certificate
	Signer          identity.SignerSerializer
	CryptoProvider  bccsp.BCCSP
}

// ClientConnectionsInput holds the input parameters for creating
// client connections.
type ClientConnectionsInput struct {
	CommandName      string
	EndorserRequired bool
	OrdererRequired  bool
	OrderingEndpoint string
	ChannelID        string
	PeerAddresses    []string
	//TLSRootCertFiles      []string
	//ConnectionProfilePath string
	TargetPeer string
	TLSEnabled bool
}

// NewClientConnections creates a new set of client connections based on the
// input parameters.
func NewClientConnections(input *ClientConnectionsInput, cryptoProvider bccsp.BCCSP) (*ClientConnections, error) {
	signer, err := common.GetDefaultSigner()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to retrieve default signer")
	}

	c := &ClientConnections{
		Signer:         signer,
		CryptoProvider: cryptoProvider,
	}

	//if input.EndorserRequired {
	//	err := c.setPeerClients(input)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	if input.OrdererRequired {
		err := c.setOrdererClient()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

//func (c *ClientConnections) setPeerClients(input *ClientConnectionsInput) error {
//	var endorserClients []pb.EndorserClient
//	var deliverClients []pb.DeliverClient
//
//	if err := c.validatePeerConnectionParameters(input); err != nil {
//		return errors.WithMessage(err, "failed to validate peer connection parameters")
//	}
//
//	for i, address := range input.PeerAddresses {
//		var tlsRootCertFile string
//		if input.TLSRootCertFiles != nil {
//			tlsRootCertFile = input.TLSRootCertFiles[i]
//		}
//		endorserClient, err := common.GetEndorserClient(address, tlsRootCertFile)
//		if err != nil {
//			return errors.WithMessagef(err, "failed to retrieve endorser client for %s", input.CommandName)
//		}
//		endorserClients = append(endorserClients, endorserClient)
//		deliverClient, err := common.GetPeerDeliverClient(address, tlsRootCertFile)
//		if err != nil {
//			return errors.WithMessagef(err, "failed to retrieve deliver client for %s", input.CommandName)
//		}
//		deliverClients = append(deliverClients, deliverClient)
//	}
//	if len(endorserClients) == 0 {
//		// this should only be empty due to a programming bug
//		return errors.New("no endorser clients retrieved")
//	}
//
//	err := c.setCertificate()
//	if err != nil {
//		return err
//	}
//
//	c.EndorserClients = endorserClients
//	c.DeliverClients = deliverClients
//
//	return nil
//}

func (c *ClientConnections) setOrdererClient() error {
	oe := viper.GetString("orderer.address")
	if oe == "" {
		// if we're here we didn't get an orderer endpoint from the command line
		// so we'll attempt to get one from cscc - bless it
		if c.Signer == nil {
			return errors.New("cannot obtain orderer endpoint, no signer was configured")
		}

		if len(c.EndorserClients) == 0 {
			return errors.New("cannot obtain orderer endpoint, empty endorser list")
		}

		orderingEndpoints, err := common.GetOrdererEndpointOfChainFnc(channelID, c.Signer, c.EndorserClients[0], c.CryptoProvider)
		if err != nil {
			return errors.WithMessagef(err, "error getting channel (%s) orderer endpoint", channelID)
		}
		if len(orderingEndpoints) == 0 {
			return errors.Errorf("no orderer endpoints retrieved for channel %s, pass orderer endpoint with -o flag instead", channelID)
		}

		logger.Infof("Retrieved channel (%s) orderer endpoint: %s", channelID, orderingEndpoints[0])
		// override viper env
		viper.Set("orderer.address", orderingEndpoints[0])
	}

	broadcastClient, err := common.GetBroadcastClient()
	if err != nil {
		return errors.WithMessage(err, "failed to retrieve broadcast client")
	}

	c.BroadcastClient = broadcastClient

	return nil
}
