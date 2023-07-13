package chaincode

import (
	"context"
	"crypto/tls"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/spf13/viper"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/internal/peer/chaincode"
	"github.com/hyperledger/fabric/internal/peer/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

// Committer holds the dependencies needed to commit
// a chaincode
type Committer struct {
	Certificate     tls.Certificate
	Command         *cobra.Command
	BroadcastClient common.BroadcastClient
	EndorserClients []EndorserClient
	DeliverClients  []pb.DeliverClient
	Signer          Signer
}

var chaincodeApproveCmd *cobra.Command

// approveForThisPeerCmd returns the cobra command for Chaincode ApproveSensoryReadingForThisPeer
func ApproveForThisPeerCmd(cryptoProvider bccsp.BCCSP) *cobra.Command {
	chaincodeApproveCmd = &cobra.Command{
		Use:   "approveForThisPeer",
		Short: "FOR INTERNAL USE ONLY. Approve a sensory reading for this peer",
		Long:  "FOR INTERNAL USE ONLY. Approve a sensory reading for this peer",
		RunE: func(cmd *cobra.Command, args []string) error {
			ccInput := &ClientConnectionsInput{
				CommandName:      cmd.Name(),
				EndorserRequired: false,
				OrdererRequired:  true,
				ChannelID:        channelID,
				PeerAddresses:    peerAddresses,
				//TLSRootCertFiles:      tlsRootCertFiles,
				//ConnectionProfilePath: connectionProfilePath,
				TLSEnabled: viper.GetBool("peer.tls.enabled"),
			}
			cc, err := NewClientConnections(ccInput, cryptoProvider)
			if err != nil {
				return err
			}

			c := &Committer{
				Command:         cmd,
				Certificate:     cc.Certificate,
				BroadcastClient: cc.BroadcastClient,
				DeliverClients:  cc.DeliverClients,
				Signer:          cc.Signer,
			}

			return c.chaincodeApprove()
		},
	}

	return chaincodeApproveCmd
}

func (c *Committer) chaincodeApprove() error {
	var err error

	if c.Command != nil {
		// Parsing of the command line is done so silence cmd usage
		c.Command.SilenceUsage = true
	}

	// assemble a signed transaction (it's an Envelope message)
	creator := protoutil.MarshalOrPanic(&msp.SerializedIdentity{
		Mspid:   "OrgDummyMSPOrderer",
		IdBytes: []byte("dummy"),
	})
	nonce := protoutil.CreateNonceOrPanic()
	txId := protoutil.ComputeTxID(nonce, creator)

	env := &cb.Envelope{
		Payload: protoutil.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: protoutil.MarshalOrPanic(&cb.ChannelHeader{
					Type:    int32(cb.HeaderType_PEER_SIGNATURE_TX),
					Version: int32(0),
					Timestamp: &timestamp.Timestamp{
						Seconds: time.Now().Unix(),
						Nanos:   0,
					},
					TxId:      txId,
					ChannelId: "mychannel",
					Epoch:     0,
				}),
				SignatureHeader: protoutil.MarshalOrPanic(&cb.SignatureHeader{
					Creator: creator,
					Nonce:   nonce,
				}),
			},
			Data: []byte{},
		}),
		Signature: []byte("signature"),
	}

	var dg *chaincode.DeliverGroup
	var ctx context.Context
	if waitForEvent {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(context.Background(), waitForEventTimeout)
		defer cancelFunc()

		dg = chaincode.NewDeliverGroup(
			c.DeliverClients,
			make([]string, len(c.DeliverClients)),
			c.Signer,
			c.Certificate,
			channelID,
			txId,
		)
		// connect to deliver service on all peers
		err := dg.Connect(ctx)
		if err != nil {
			return err
		}
	}

	if err = c.BroadcastClient.Send(env); err != nil {
		return errors.WithMessage(err, "failed to send transaction")
	}

	if dg != nil && ctx != nil {
		// wait for event that contains the txID from all peers
		err = dg.Wait(ctx)
		if err != nil {
			return err
		}
	}
	return err
}
