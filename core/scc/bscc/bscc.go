package bscc

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	event "github.com/hyperledger/fabric/common/blocc-events"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/internal/peer/common"
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
		CryptoProvider:      bscc.peerInstance.CryptoProvider,
	}
	bloccProtoLogger.Infof("BSCC config: %+v", bscc.config)

	return shim.Success(nil)
}

func (bscc *BSCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("[BLOCC System CC] This invoke function is not allowed for external calls," +
		" only internal calls are allowed.")
}

// ----------------- BSCC Implementation ----------------- //

// TODO: see if it's worth checking if the signer is correct
func (bscc *BSCC) sendApprovalToOrderer(channelID string, SensoryTxID []byte) pb.Response {
	bloccProtoLogger.Debug("sendApprovalToOrderer")
	var err error

	signer, _ := common.GetDefaultSigner()
	env, txID, envErr := protoutil.CreateSignedEnvelopeWithTxID(
		cb.HeaderType_PEER_SIGNATURE_TX,
		channelID,
		signer,
		&cb.ApprovalTxEnvelope{TxId: SensoryTxID},
		0,
		0,
	)
	if envErr != nil {
		bloccProtoLogger.Fatal("Error creating signed envelope: " + envErr.Error())
	}

	dg, ctx := bscc.createDeliverGroupAndContext(txID)

	if dg == nil && ctx == nil {
		return shim.Error("Failed to create deliver group or context.")
	}

	broadcastClient, err := createBroadcastClient(&bscc.config)

	if err != nil {
		return shim.Error("Failed to retrieve broadcast client: " + err.Error())
	}
	if err = broadcastClient.Send(env); err != nil {
		return shim.Error("Failed to send transaction: " + err.Error())
	}

	if dg != nil && ctx != nil {
		// wait for event that contains the txID from all peers
		err = dg.Wait(ctx)
		if err != nil {
			bloccProtoLogger.Errorf("Failed to wait for event: %s", err)
			return shim.Error("Failed to wait for event: " + err.Error())
		}
	}
	return shim.Success(nil)
}

//func (bscc *BSCC) sendApprovalToOrderer2(channelID string) pb.Response {
//	bloccProtoLogger.Debug("sendApprovalToOrderer")
//	var err error
//
//	signer, _ := common.GetDefaultSigner()
//
//	proposal, txID, err := bscc.createProposal(channelID, txID, signer)
//	if err != nil {
//		return shim.Error("failed to create proposal: " + err.Error())
//	}
//
//	signedProposal, err := bscc.signProposal(proposal, signer)
//	if err != nil {
//		return shim.Error("failed to sign proposal: " + err.Error())
//	}
//
//	var responses []*pb.ProposalResponse
//	for _, endorser := range c.EndorserClients {
//		proposalResponse, err := endorser.ProcessProposal(context.Background(), signedProposal)
//		if err != nil {
//			return errors.WithMessage(err, "failed to endorse proposal")
//		}
//		responses = append(responses, proposalResponse)
//	}
//
//	if len(responses) == 0 {
//		// this should only be empty due to a programming bug
//		return errors.New("no proposal responses received")
//	}
//
//	// all responses will be checked when the signed transaction is created.
//	// for now, just set this so we check the first response's status
//	proposalResponse := responses[0]
//
//	if proposalResponse == nil {
//		return errors.New("received nil proposal response")
//	}
//
//	if proposalResponse.Response == nil {
//		return errors.New("received proposal response with nil response")
//	}
//
//	if proposalResponse.Response.Status != int32(cb.Status_SUCCESS) {
//		return errors.Errorf("proposal failed with status: %d - %s", proposalResponse.Response.Status, proposalResponse.Response.Message)
//	}
//	// assemble a signed transaction (it's an Envelope message)
//	env, err := protoutil.CreateSignedTx(proposal, c.Signer, responses...)
//	if err != nil {
//		return errors.WithMessage(err, "failed to create signed transaction")
//	}
//
//	dg, ctx := bscc.createDeliverGroupAndContext(txID)
//
//	if dg == nil && ctx == nil {
//		return shim.Error("Failed to create deliver group or context.")
//	}
//
//	broadcastClient, err := createBroadcastClient(&bscc.config)
//
//	if err != nil {
//		return shim.Error("Failed to retrieve broadcast client: " + err.Error())
//	}
//	if err = broadcastClient.Send(env); err != nil {
//		return shim.Error("Failed to send transaction: " + err.Error())
//	}
//
//	if dg != nil && ctx != nil {
//		// wait for event that contains the txID from all peers
//		err = dg.Wait(ctx)
//		if err != nil {
//			bloccProtoLogger.Errorf("Failed to wait for event: %s", err)
//			return shim.Error("Failed to wait for event: " + err.Error())
//		}
//	}
//	return shim.Success(nil)
//}
//
//func (bscc *BSCC) createProposal(channelID string, inputTxID string, signer msp.SigningIdentity) (proposal *pb.Proposal, txID string, err error) {
//	args := &lb.CommitChaincodeDefinitionArgs{
//		Name:              "bscc",
//		Version:           "1",
//		Sequence:          1,
//		EndorsementPlugin: "escc",
//		ValidationPlugin:  "vscc",
//	}
//
//	argsBytes, err := proto.Marshal(args)
//	if err != nil {
//		return nil, "", err
//	}
//	ccInput := &pb.ChaincodeInput{Args: [][]byte{[]byte("")}, argsBytes}
//
//	cis := &pb.ChaincodeInvocationSpec{
//		ChaincodeSpec: &pb.ChaincodeSpec{
//			ChaincodeId: &pb.ChaincodeID{Name: bscc.Name()},
//			Input:       ccInput,
//		},
//	}
//
//	creatorBytes, err := signer.Serialize()
//	if err != nil {
//		return nil, "", errors.WithMessage(err, "failed to serialize identity")
//	}
//
//	proposal, txID, err = protoutil.CreateChaincodeProposalWithTxIDAndTransient(cb.HeaderType_ENDORSER_TRANSACTION, channelID, cis, creatorBytes, inputTxID, nil)
//	if err != nil {
//		return nil, "", errors.WithMessage(err, "failed to create ChaincodeInvocationSpec proposal")
//	}
//
//	return proposal, txID, nil
//}
//
//func (bscc *BSCC) signProposal(proposal *pb.Proposal, signer msp.SigningIdentity) (*pb.SignedProposal, error) {
//	// check for nil argument
//	if proposal == nil {
//		return nil, errors.New("proposal cannot be nil")
//	}
//
//	if signer == nil {
//		return nil, errors.New("signer cannot be nil")
//	}
//
//	proposalBytes, err := proto.Marshal(proposal)
//	if err != nil {
//		return nil, errors.Wrap(err, "error marshaling proposal")
//	}
//
//	signature, err := signer.Sign(proposalBytes)
//	if err != nil {
//		return nil, err
//	}
//
//	return &pb.SignedProposal{
//		ProposalBytes: proposalBytes,
//		Signature:     signature,
//	}, nil
//}

func (bscc *BSCC) processEvent(event event.Event) {
	bloccProtoLogger.Info("BLOCC - Received approval event: ", event)
	bscc.sendApprovalToOrderer(event.ChannelID, event.SensoryTxID)
}
