package blocc

import (
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/internal/peer/blocc/chaincode"
	"github.com/spf13/cobra"
)

// Cmd returns the cobra command for lifecycle
func Cmd(cryptoProvider bccsp.BCCSP) *cobra.Command {
	bloccCmd := &cobra.Command{
		Use:   "blocc",
		Short: "Perform blocc operations",
		Long:  "Perform blocc operations",
	}
	bloccCmd.AddCommand(chaincode.Cmd(cryptoProvider))

	return bloccCmd
}
