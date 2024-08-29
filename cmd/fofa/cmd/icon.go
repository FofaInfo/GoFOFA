package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	openBrowser bool
)

// icon subcommand
var iconCmd = &cli.Command{
	Name:                   "icon",
	Usage:                  "fofa icon search",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "open",
			Value:       false,
			Usage:       "open fofa website once find the favicon",
			Destination: &openBrowser,
		},
	},
	Action: iconAction,
}

// openURL open browser to url
//func openURL(url string) error {
//	var err error
//
//	switch runtime.GOOS {
//	case "linux":
//		err = exec.Command("xdg-open", url).Start()
//	case "windows":
//		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
//	case "darwin":
//		err = exec.Command("open", url).Start()
//	default:
//		err = fmt.Errorf("unsupported platform")
//	}
//
//	return err
//}

// iconAction icon action
// url can be: local file; remote favicon url; remote homepage;
func iconAction(ctx *cli.Context) error {
	// valid same config
	url := ctx.Args().First()
	if len(url) == 0 {
		return errors.New("url cannot be empty")
	}

	logrus.Debug("open url: ", url)

	// do search
	hash, err := gofofa.IconHash(url)
	if err != nil {
		return err
	}

	fmt.Println(hash)

	if openBrowser {
		openURL := fofaCli.Server + "/result?qbase64=" + base64.StdEncoding.EncodeToString([]byte("icon_hash="+hash))
		logrus.Debug("open fofa query browser: ", openURL)
		if err = browser.OpenURL(openURL); err != nil {
			return err
		}

		//if err = openURL(fofaCli.Server + "/result?qbase64=" + base64.StdEncoding.EncodeToString([]byte(hash))); err != nil {
		//	return err
		//}
	}

	return nil
}
