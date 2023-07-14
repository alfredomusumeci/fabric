package protoutil

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
)

// CreateSignedEnvelopeWithTxID creates a signed envelope of the desired type, with
// marshaled dataMsg and signs it
func CreateSignedEnvelopeWithTxID(
	txType common.HeaderType,
	channelID string,
	signer Signer,
	dataMsg proto.Message,
	msgVersion int32,
	epoch uint64,
) (*common.Envelope, string, error) {
	return CreateSignedEnvelopeWithTxIDWithTLSBinding(txType, channelID, signer, dataMsg, msgVersion, epoch, nil)
}

// CreateSignedEnvelopeWithTxIDWithTLSBinding creates a signed envelope of the desired
// type, with marshaled dataMsg and signs it. It also includes a TLS cert hash
// into the channel header
func CreateSignedEnvelopeWithTxIDWithTLSBinding(
	txType common.HeaderType,
	channelID string,
	signer Signer,
	dataMsg proto.Message,
	msgVersion int32,
	epoch uint64,
	tlsCertHash []byte,
) (*common.Envelope, string, error) {
	payloadChannelHeader := MakeChannelHeader(txType, msgVersion, channelID, epoch)
	payloadChannelHeader.TlsCertHash = tlsCertHash
	var err error
	payloadSignatureHeader := &common.SignatureHeader{}
	SetTxID(payloadChannelHeader, payloadSignatureHeader)

	if signer != nil {
		payloadSignatureHeader, err = NewSignatureHeader(signer)
		if err != nil {
			return nil, "", err
		}
	}

	data, err := proto.Marshal(dataMsg)
	if err != nil {
		return nil, "", errors.Wrap(err, "error marshaling")
	}

	paylBytes := MarshalOrPanic(
		&common.Payload{
			Header: MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
			Data:   data,
		},
	)

	var sig []byte
	if signer != nil {
		sig, err = signer.Sign(paylBytes)
		if err != nil {
			return nil, "", err
		}
	}

	env := &common.Envelope{
		Payload:   paylBytes,
		Signature: sig,
	}
	txID := payloadChannelHeader.TxId

	return env, txID, nil
}
