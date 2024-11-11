package cmd

import (
	"os"

	"github.com/LubyRuffy/gofofa"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	fofaCli *gofofa.Client
)

var (
	fofaURL string // fofa url
)

// GlobalCommands global commands
var GlobalCommands = []*cli.Command{
	searchCmd,
	accountCmd,
	countCmd,
	statsCmd,
	iconCmd,
	randomCmd,
	hostCmd,
	dumpCmd,
	domainsCmd,
	activeCmd,
	dedupCmd,
	categoryCmd,
	browserCmd,
}

// IsValidCommand valid command name
func IsValidCommand(cmd string) bool {
	if len(cmd) == 0 {
		return false
	}
	if cmd[0] == '-' {
		switch cmd {
		case "--help", "-help", "-h", "--version", "-version", "-v":
			// 自带的配置
			return true
		default:
			for _, option := range GlobalOptions {
				for _, name := range option.Names() {
					if cmd == "--"+name || cmd == "-"+name {
						return true
					}
				}
			}
		}
		return false
	}

	for _, command := range GlobalCommands {
		if command.Name == cmd {
			return true
		}
	}
	return false
}

// GlobalOptions global options
var GlobalOptions = []cli.Flag{
	&cli.StringFlag{
		Name:        "fofaURL",
		Aliases:     []string{"u"},
		Value:       gofofa.FofaURLFromEnv(),
		Usage:       "format: <url>/?email=<email>&key=<key>&version=<v2>",
		Destination: &fofaURL,
	},
	&cli.BoolFlag{
		Name:  "verbose",
		Usage: "print more information",
	},
	&cli.BoolFlag{
		Name:  "accountDebug",
		Usage: "print account in error log",
	},
}

//// isSubCmd 判断是否指定的子命令
//func isSubCmd(args []string, cmd string) bool {
//	for _, arg := range args {
//		if arg[0] == '-' {
//			continue
//		}
//		if strings.Contains(arg, "=") {
//			continue
//		}
//		if arg == cmd {
//			return true
//		}
//	}
//	return false
//}

// BeforAction generate fofa client
func BeforAction(context *cli.Context) error {
	var err error

	// not any command, and no query for default command(search)
	if len(os.Args) == 1 {
		return nil
	}

	//logrus.SetOutput(os.Stderr) // 默认
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	if context.Bool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	var accountDebug bool
	if context.Bool("accountDebug") {
		accountDebug = true
	}

	//// icon no need client
	//if isSubCmd(os.Args[1:], "icon") {
	//	return nil
	//}

	fofaCli, err = gofofa.NewClient(gofofa.WithURL(fofaURL), gofofa.WithAccountDebug(accountDebug))
	if err != nil {
		return err
	}

	if len(deductMode) > 0 {
		fofaCli.DeductMode = gofofa.ParseDeductMode(deductMode)
	}

	return nil
}
