package start

import (
	"crawlab/app/master/config"
	"crawlab/pkg/core/cli/cliext"
	"crawlab/utils"
	"errors"
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func NewCommand() *cli.Command {
	var prefix string
	var prefixFlag = &cli.StringFlag{
		Name:        "env_prefix",
		Value:       "CRAWLAB",
		Destination: &prefix,
		DefaultText: "CRAWLAB",
	}
	var configFlag = &cli.PathFlag{
		Name:     "config",
		Required: true,
	}
	flags := []cli.Flag{
		configFlag,
		prefixFlag,
	}
	err := prefixFlag.Apply(flag.CommandLine)
	if err == nil {
		err = cliext.GenerateCliFlags(config.DefaultConfig, prefix, "", &flags)
	}
	conf := new(config.ApplicationConfig)
	return &cli.Command{
		Name:  "start",
		Flags: flags,
		Before: func(context *cli.Context) error {
			if err != nil {
				return err
			}
			if !utils.Exists(context.Path("config")) {
				return errors.New("配置文件不存在")
			}
			err = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc(configFlag.Name))(context)
			if err != nil {
				return err
			}
			err = cliext.DecodeCliFlagsTo(context, "", conf)
			if err != nil {
				return err
			}
			return nil
		},
		Action: func(context *cli.Context) error {

			spew.Dump(context.App.VisibleFlags())
			return nil
			//bt, err := server.Launcher(conf)
			//if err != nil {
			//	return err
			//}
			//return bt.Run()
		},
	}
}
