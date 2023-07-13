package chaincode

import (
	"time"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/peer/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	bloccName       = "_blocc"
	approveFuncName = "ApproveSensoryReadingForThisPeer"
	infoFuncName    = "Usage"
)

var logger = flogging.MustGetLogger("cli.blocc.chaincode")

// Cmd returns the cobra command for Chaincode
func Cmd(cryptoProvider bccsp.BCCSP) *cobra.Command {
	chaincodeCmd.PersistentFlags()

	//chaincodeCmd.AddCommand(InfoCmd(nil, cryptoProvider))
	chaincodeCmd.AddCommand(ApproveForThisPeerCmd(cryptoProvider))

	return chaincodeCmd
}

var (
	channelID           string
	peerAddresses       []string
	waitForEvent        bool
	waitForEventTimeout time.Duration
)

var chaincodeCmd = &cobra.Command{
	Use:   "_blocc",
	Short: "Perform _blocc protocol related operations, not intended to be called directly by end users",
	Long:  "Perform _blocc protocol related operations, not intended to be called directly by end users",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		common.InitCmd(cmd, args)
	},
}

var flags *pflag.FlagSet

func init() {
	ResetFlags()
}

// ResetFlags resets the values of these flags to facilitate tests
func ResetFlags() {
	flags = &pflag.FlagSet{}

	flags.StringVarP(&channelID, "channelID", "c", "", "The channel on which this command should be executed")
	flags.StringArrayVarP(&peerAddresses, "peerAddresses", "", []string{""}, "The addresses of the peers to connect to")
	flags.BoolVar(&waitForEvent, "waitForEvent", true,
		"Whether to wait for the event from each peer's deliver filtered service signifying that the transaction has been committed successfully")
	flags.DurationVar(&waitForEventTimeout, "waitForEventTimeout", 30*time.Second,
		"Time to wait for the event from each peer's deliver filtered service signifying that the 'invoke' transaction has been committed successfully")
}

func attachFlags(cmd *cobra.Command, names []string) {
	cmdFlags := cmd.Flags()
	for _, name := range names {
		if flag := flags.Lookup(name); flag != nil {
			cmdFlags.AddFlag(flag)
		} else {
			logger.Fatalf("Could not find flag '%s' to attach to command '%s'", name, cmd.Name())
		}
	}
}
