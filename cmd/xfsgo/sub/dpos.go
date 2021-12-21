package sub

import (
	"fmt"
	"xfsgo"
	"xfsgo/common"

	"github.com/spf13/cobra"
)

// var fileType string = ".car"
var (
	dposCommand = &cobra.Command{
		Use:                   "dpos <command> [options]",
		DisableFlagsInUseLine: true,
		Short:                 "show dpos info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	getValidatorsCommond = &cobra.Command{
		Use:                   "validators [options] <number>",
		DisableFlagsInUseLine: true,
		Short:                 "GetValidators retrieves the list of the validators at specified block",
		RunE:                  getValidators,
	}

	getConfirmedBlockByNumCommond = &cobra.Command{
		Use:                   "getConfirmed [options] <number>",
		DisableFlagsInUseLine: true,
		Short:                 "GetConfirmedBlockNumber retrieves the latest irreversible block",
		RunE:                  getConfirmedBlockNumber,
	}
)

func getValidators(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	req := &GetBlockNumByValidatorArgs{
		Number: args[0],
	}

	validators := make([]common.Address, 0)
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "Dpos.GetValidators", &req, &validators)
	if err != nil {
		return err
	}

	bs, err := common.MarshalIndent(validators)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", string(bs))
	return nil
}

func getConfirmedBlockNumber(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}

	var result int
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "Dpos.GetConfirmedBlockNumber", nil, &result)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	return nil
}

func init() {

	rootCmd.AddCommand(dposCommand)
	dposCommand.AddCommand(getValidatorsCommond)
	dposCommand.AddCommand(getConfirmedBlockByNumCommond)
}
