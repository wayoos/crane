package main

import (
	"container/list"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	"github.com/wayoos/crane/api/client"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/api/server"
	"os"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "crane"
	app.Usage = "crane [command]"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Value: "http://localhost:2475",
			Usage: "Host",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "lup",
			Usage: "Create and start container",
			//			Description: "Create and start container",
			Action: client.UpCommand,
		},
		{
			Name:  "up",
			Usage: "Create and start container",
			//			Description: "Create and start container",
			Action: client.UpCommand,
		},
		{
			Name:        "rm",
			Usage:       "crane rm",
			Description: "Stop and remove container",
			Action:      client.RmCommand,
		},
		{
			Name:        "push",
			ShortName:   "p",
			Usage:       "crane push PATH",
			Description: "push an image package or a crane package to the crane server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "",
					Usage: "Load name (and optionally a tag) to be applied to the resulting",
				},
			},
			Action: client.PushCommand,
		},
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "start crane server",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "p, port",
					Value: 2475,
					Usage: "port to listen on (default 2475)",
				},
				cli.StringFlag{
					Name:  "c, crane-dir",
					Value: "current directory",
					Usage: "Path to use as the crane data directory",
				},
			},
			Action: server.ServerCommand,
		},
		{
			Name:      "exec",
			ShortName: "e",
			Usage:     "crane exec LOADID command...",
			Action: func(c *cli.Context) {
				loadId := c.Args().First()

				//fmt.Println("Execute cmd in " + loadId)

				host := c.GlobalString("host")

				var cmds []string = c.Args().Tail()

				l := list.New()

				for i := range cmds {
					//					println(cmds[i])

					val := cmds[i]
					split := strings.Split(val, " ")
					for si := range split {
						l.PushBack(split[si])
					}
				}

				cmds = make([]string, l.Len())

				idx := 0
				for e := l.Front(); e != nil; e = e.Next() {
					cmd := e.Value.(string)
					cmds[idx] = cmd
					idx++
				}

				//				for i := range cmds {
				//					println(cmds[i])
				//				}

				execData := domain.ExecData{
					LoadId: loadId,
					Cmd:    cmds,
				}

				result := domain.ExecResult{}
				resp, err := napping.Post(host+"/exec", &execData, &result, nil)
				if err != nil {
					panic(err)
				}
				if resp.Status() == 200 {
					fmt.Println(result.Out)
				}

			},
		},
		{
			Name:   "ps",
			Usage:  "crane ps",
			Action: client.PsCommand,
		},
	}

	app.Action = func(c *cli.Context) {

		if len(c.Args()) <= 0 {
			cli.ShowAppHelp(c)
		} else {

			println("Invalid command")
		}

	}

	app.Run(os.Args)
}
