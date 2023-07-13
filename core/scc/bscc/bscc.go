package bscc

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	event "github.com/hyperledger/fabric/common/blocc-events"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/internal/peer/chaincode"
	"github.com/hyperledger/fabric/internal/peer/common"
	//id "github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protoutil"
	"os"
	"time"
)

func New(peerInstance *peer.Peer) *BSCC {
	return &BSCC{
		peerInstance: peerInstance,
	}
}

func (bscc *BSCC) Name() string {
	return "bscc"
}

func (bscc *BSCC) Chaincode() shim.Chaincode {
	return bscc
}

type BSCC struct {
	peerInstance *peer.Peer
	config       Config
}

var bloccProtoLogger = flogging.MustGetLogger("bscc")

const (
	ApproveSensoryReading string = "ApproveForThisPeer"
)

// ------------------- Error handling ------------------- //

type InvalidFunctionError string

func (f InvalidFunctionError) Error() string {
	return fmt.Sprintf("invalid function to bscc: %s", string(f))
}

// -------------------- Stub Interface ------------------- //

func (bscc *BSCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	bloccProtoLogger.Info("Init BSCC")
	go func() {
		for _event := range event.GlobalEventBus.Subscribe() {
			bscc.processEvent(_event)
		}
	}()

	peerAddress, ok := os.LookupEnv("CORE_PEER_ADDRESS")
	if !ok {
		bloccProtoLogger.Error("CORE_PEER_ADDRESS is not set")
		return shim.Error("CORE_PEER_ADDRESS is not set")
	}

	tlsCertFile, ok := os.LookupEnv("CORE_PEER_TLS_ROOTCERT_FILE")
	if !ok {
		bloccProtoLogger.Error("CORE_PEER_TLS_ROOTCERT_FILE is not set")
		return shim.Error("CORE_PEER_TLS_ROOTCERT_FILE is not set")
	}

	orgMspID, ok := os.LookupEnv("CORE_PEER_LOCALMSPID")
	if !ok {
		bloccProtoLogger.Error("CORE_PEER_LOCALMSPID is not set")
		return shim.Error("CORE_PEER_LOCALMSPID is not set")
	}

	bscc.config = Config{
		PeerAddress:         peerAddress,
		TLSCertFile:         tlsCertFile,
		OrgMspID:            orgMspID,
		WaitForEvent:        true,
		WaitForEventTimeout: 3 * time.Second,
	}
	bloccProtoLogger.Infof("BSCC config: %+v", bscc.config)

	return shim.Success(nil)
}

func (bscc *BSCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetArgs()
	if len(args) < 1 {
		return shim.Error("invalid args")
	}

	function := string(args[0])

	bloccProtoLogger.Debugf("Invoke BSCC function: %s", function)

	//TODO: do I need to handle ACL?

	switch function {
	case ApproveSensoryReading:
		// TODO: mspID is a string and idBytes is a byte array. Need to convert them to the right types.
		// Also technically this function is never used, so it's not a priority.
		// Check for the right number of args
		if len(args) != 3 {
			return shim.Error("incorrect number of arguments. Expecting 3")
		}

		mspID := string(args[1])
		idBytes := args[2]

		response := bscc.sendApprovalToOrderer(mspID, idBytes)
		if response.Status != shim.OK {
			return shim.Error("Approval block was not uploaded due to: " + response.GetMessage())
		}
		return shim.Success([]byte("OK"))
	}

	return shim.Error(InvalidFunctionError(function).Error())
}

// ----------------- BSCC Implementation ----------------- //

func (bscc *BSCC) sendApprovalToOrderer(mspId string, idBytes []byte) pb.Response {
	bloccProtoLogger.Debug("sendApprovalToOrderer")
	var err error
	bloccProtoLogger.Debug("mspId: " + mspId)
	bloccProtoLogger.Debug("idBytes: " + string(idBytes))
	//assemble a signed transaction (it's an Envelope message)
	creator := protoutil.MarshalOrPanic(&msp.SerializedIdentity{
		Mspid:   mspId,
		IdBytes: idBytes,
	})
	//serializedIdentity, err := approver.Serialize()
	//if err != nil {
	//	fmt.Printf("Error serializing the signing identity: %v\n", err)
	//	return
	//}

	nonce := protoutil.CreateNonceOrPanic()
	txId := protoutil.ComputeTxID(nonce, creator)
	signer, _ := common.GetDefaultSigner()
	bloccProtoLogger.Debug("signer: " + signer.GetMSPIdentifier())
	env, envErr := protoutil.CreateSignedEnvelope(cb.HeaderType_PEER_SIGNATURE_TX, "mychannel", signer, &cb.Envelope{}, 0, 0)
	if envErr != nil {
		bloccProtoLogger.Fatal("Error creating signed envelope: " + envErr.Error())
	}

	//env := &cb.Envelope{
	//	Payload: protoutil.MarshalOrPanic(&cb.Payload{
	//		Header: &cb.Header{
	//			ChannelHeader: protoutil.MarshalOrPanic(&cb.ChannelHeader{
	//				Type:    int32(cb.HeaderType_PEER_SIGNATURE_TX),
	//				Version: int32(0),
	//				Timestamp: &timestamp.Timestamp{
	//					Seconds: time.Now().Unix(),
	//					Nanos:   0,
	//				},
	//				TxId:      txId,
	//				ChannelId: "mychannel",
	//				Epoch:     0,
	//			}),
	//			SignatureHeader: protoutil.MarshalOrPanic(&cb.SignatureHeader{
	//				Creator: creator,
	//				Nonce:   nonce,
	//			}),
	//		},
	//		Data: []byte{},
	//	}),
	//	Signature: signature,
	//}
	//From the below, clientCertificate can be obtianed

	var dg *chaincode.DeliverGroup
	var ctx context.Context
	if bscc.config.WaitForEvent {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(context.Background(), bscc.config.WaitForEventTimeout)
		defer cancelFunc()

		dg, err = createDeliverGroup(&bscc.config, txId)
		if err != nil {
			bloccProtoLogger.Errorf("Failed to create deliver group: %s", err)
		}
		bloccProtoLogger.Debug("Created deliver group")
		// connect to deliver service on all peers
		err = dg.Connect(ctx)
		if err != nil {
			bloccProtoLogger.Errorf("Failed to connect to deliver service: %s", err)
			return shim.Error("Failed to connect to deliver service: " + err.Error())
		}

	}

	bloccProtoLogger.Debug("Sending transaction to orderer with id %s", txId)
	broadcastClient, err := createBroadcastClient(&bscc.config)
	bloccProtoLogger.Debugf("Got broadcast client: %v", broadcastClient)

	if err != nil {
		bloccProtoLogger.Errorf("Failed to retrieve broadcast client: %s", err)
		return shim.Error("Failed to retrieve broadcast client: " + err.Error())
	}
	if err = broadcastClient.Send(env); err != nil {
		bloccProtoLogger.Errorf("Failed to send transaction: %s", err)
		return shim.Error("Failed to send transaction: " + err.Error())
	}
	bloccProtoLogger.Debug("Transaction sent to orderer")

	if dg != nil && ctx != nil {
		// wait for event that contains the txID from all peers
		err = dg.Wait(ctx)
		if err != nil {
			bloccProtoLogger.Errorf("Failed to wait for event: %s", err)
			return shim.Error("Failed to wait for event: " + err.Error())
		}
	}
	bloccProtoLogger.Debug("End of sendApprovalToOrderer")
	return shim.Success(nil)
}

func (bscc *BSCC) processEvent(event event.Event) {
	bloccProtoLogger.Info("Received event: ", event)

	bscc.sendApprovalToOrderer(event.MspID, event.IdBytes)
}
