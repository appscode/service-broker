package cmds

import (
	"io"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/cli"
	"github.com/kubedb/redis/pkg/cmds/server"
	"github.com/spf13/cobra"
)

func NewCmdRun(version string, out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := server.NewRedisServerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch Redis operator",
		DisableAutoGenTag: true,
		PreRun: func(c *cobra.Command, args []string) {
			cli.SendPeriodicAnalytics(c, version)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infoln("Starting redis operator...")

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
