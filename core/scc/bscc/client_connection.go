package bscc

import (
	"context"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/internal/peer/chaincode"
	"github.com/hyperledger/fabric/internal/peer/common"
	"github.com/hyperledger/fabric/internal/pkg/comm"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	PeerAddress         string
	TLSCertFile         string
	OrgMspID            string
	WaitForEvent        bool
	WaitForEventTimeout time.Duration
	CryptoProvider      bccsp.BCCSP
}

//type Clients struct {
//	EndorserClients []pb.EndorserClient
//	DeliverClients  []pb.DeliverClient
//}

func (bscc *BSCC) createDeliverGroupAndContext(txID string) (*chaincode.DeliverGroup, context.Context) {
	var dg *chaincode.DeliverGroup
	var ctx context.Context
	if bscc.config.WaitForEvent {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(context.Background(), bscc.config.WaitForEventTimeout)
		defer cancelFunc()

		var err error
		dg, err = createDeliverGroup(&bscc.config, txID)
		if err != nil {
			bloccProtoLogger.Errorf("Failed to create deliver group: %s", err)
		}
		bloccProtoLogger.Debug("Created deliver group")
		// connect to deliver service on all peers
		err = dg.Connect(ctx)
		if err != nil {
			return nil, nil
		}
	}
	return dg, ctx
}

// TODO: this is limited to one peer for now, extend this to multiple
// TODO: this is also limited to one channel for now, extend this to multiple
func createDeliverGroup(config *Config, txID string) (*chaincode.DeliverGroup, error) {
	bloccProtoLogger.Debug("Creating deliver group")
	var deliverClients []pb.DeliverClient
	var peerAddresses []string

	deliverClient, err := common.GetPeerDeliverClientFnc(config.PeerAddress, config.TLSCertFile)
	if err != nil {
		err = errors.WithMessage(err, "error getting deliver client")
		return nil, err
	}
	deliverClients = append(deliverClients, deliverClient)
	bloccProtoLogger.Debug("Got deliver client: ", deliverClient)
	bloccProtoLogger.Debug("Got deliver clients: ", deliverClients)

	certificate, err := common.GetClientCertificateFnc()
	if err != nil {
		err = errors.WithMessage(err, "error getting client certificate")
		return nil, err
	}
	bloccProtoLogger.Debug("Got certificate: ", certificate)

	signer, err := common.GetDefaultSignerFnc()
	if err != nil {
		err = errors.WithMessage(err, "error getting default signer")
		return nil, err
	}
	bloccProtoLogger.Debug("Got signer: ", signer)

	peerAddresses = append(peerAddresses, config.PeerAddress)
	bloccProtoLogger.Debug("Got peer addresses: ", peerAddresses)

	dg := chaincode.NewDeliverGroup(
		deliverClients,
		peerAddresses,
		signer,
		certificate,
		"mychannel",
		txID,
	)
	bloccProtoLogger.Debug("Returning deliver group")
	return dg, nil
}

func createBroadcastClient(config *Config) (common.BroadcastClient, error) {
	endorserClient, err := common.GetEndorserClientFnc(config.PeerAddress, config.TLSCertFile)
	if err != nil {
		return nil, errors.WithMessagef(err, "error getting endorser client for bscc chaincode")
	}
	endorserClients := []pb.EndorserClient{endorserClient}

	signer, err := common.GetDefaultSignerFnc()
	if err != nil {
		err = errors.WithMessage(err, "error getting default signer")
		return nil, err
	}

	var broadcastClient common.BroadcastClient
	bloccProtoLogger.Debug("Len of ordering endpoint: ", len(common.OrderingEndpoint))
	if len(common.OrderingEndpoint) == 0 {
		if len(endorserClients) == 0 {
			return nil, errors.New("orderer is required, but no ordering endpoint or endorser client supplied")
		}
		endorserClient := endorserClients[0]

		orderingEndpoints, err := common.GetOrdererEndpointOfChainFnc("mychannel", signer, endorserClient, config.CryptoProvider)
		if err != nil {
			return nil, errors.WithMessagef(err, "error getting orderer endpoint")
		}
		if len(orderingEndpoints) == 0 {
			return nil, errors.Errorf("no orderer endpoints retrieved for channel %s, pass orderer endpoint with -o flag instead", "mychannel")
		}
		bloccProtoLogger.Infof("Retrieved channel (%s) orderer endpoint: %s", "mychannel", orderingEndpoints[0])
		// override viper env
		viper.Set("orderer.address", orderingEndpoints[0])
		viper.Set("orderer.tls.enabled", true)
	}
	address, clientConfig, err := configOrdererSettings()

	bloccProtoLogger.Debug("Getting broadcast client")
	broadcastClient, err = common.GetBroadcastClientWithParams(address, clientConfig, err)
	if err != nil {
		return nil, errors.WithMessage(err, "error getting broadcast client")
	}
	bloccProtoLogger.Debug("BLOCC: got broadcast client: ", broadcastClient)

	return broadcastClient, nil
}

// TODO: this is currently hardcoded to use the orderer.example.com:7050, ideally this should be handled by the cli
func configOrdererSettings() (address string, clientConfig comm.ClientConfig, err error) {
	address = "orderer.example.com:7050"
	clientConfig = comm.ClientConfig{}
	connTimeout := 3 * time.Second
	//if connTimeout == time.Duration(0) {
	//	connTimeout = defaultConnTimeout
	//}
	clientConfig.DialTimeout = connTimeout
	secOpts := comm.SecureOptions{
		UseTLS:             true,
		RequireClientCert:  false,
		TimeShift:          0,
		ServerNameOverride: "orderer.example.com",
	}

	if secOpts.UseTLS {
		// Print current PWD
		pwd, err1 := os.Getwd()
		if err1 != nil {
			bloccProtoLogger.Debug("Error getting PWD: ", err1)
		}
		bloccProtoLogger.Debug("PWD: ", pwd)

		tlsPath := "/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
		caPEM, res := ioutil.ReadFile(tlsPath)
		if res != nil {
			err = errors.WithMessagef(res, "unable to load tls.rootcert.file")
			return
		}
		secOpts.ServerRootCAs = [][]byte{caPEM}
	}
	//if secOpts.RequireClientCert {
	//	secOpts.Key, secOpts.Certificate, err = getClientAuthInfoFromEnv(prefix)
	//	if err != nil {
	//		return
	//	}
	//}
	clientConfig.SecOpts = secOpts
	clientConfig.MaxRecvMsgSize = comm.DefaultMaxRecvMsgSize
	//if viper.IsSet(prefix + ".maxRecvMsgSize") {
	//	clientConfig.MaxRecvMsgSize = int(viper.GetInt32(prefix + ".maxRecvMsgSize"))
	//}
	clientConfig.MaxSendMsgSize = comm.DefaultMaxSendMsgSize
	//if viper.IsSet(prefix + ".maxSendMsgSize") {
	//	clientConfig.MaxSendMsgSize = int(viper.GetInt32(prefix + ".maxSendMsgSize"))
	//}
	return
}
