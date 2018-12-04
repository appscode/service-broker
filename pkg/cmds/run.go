package cmds

import (
	"github.com/appscode/service-broker/pkg/cmds/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	o := server.NewBrokerServerOptions()

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch AppsCode Service Broker server",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			glog.Infoln("Starting service broker server...")

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
