package main

import (
	"github.com/codegangsta/cli"
	"github.com/wayoos/crane/api/client"
	"github.com/wayoos/crane/api/server"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// TODO add this when crane server is disable
// quimbaya:wayoos steph$ docker ps
// FATA[0000] Cannot connect to the Docker daemon. Is 'docker -d' running on this host?

func main2() {
	cmd := exec.Command("docker", "build")

	//	if err := cmd.Start(); err != nil {
	//		log.Fatalf("cmd.Start: %v")
	//	}

	//	cmd.Wait()

	_, err := cmd.CombinedOutput()

	if err != nil {
		log.Println("err not nil")
		log.Print(err)
	}

	if !cmd.ProcessState.Success() {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)

		log.Printf("Exit Status: %d", status.ExitStatus())
	} else {
		log.Println("Success")
	}

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)

	log.Printf("Exit Status: %d", status.ExitStatus())

}

func main() {
	app := cli.NewApp()
	app.Name = "crane"
	app.Usage = "crane [command]"
	app.Version = "0.0.2"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Value: "http://localhost:2475",
			Usage: "Host",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "build",
			ShortName:   "b",
			Usage:       "Create container",
			Description: "push an image package or a crane package to the crane server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "",
					Usage: "Load name (and optionally a tag) to be applied to the resulting",
				},
			},
			Action: client.BuildCommand,
		},
		{
			Name:   "ps",
			Usage:  "List containers",
			Action: client.PsCommand,
		},
		{
			Name:        "rm",
			Usage:       "crane rm",
			Description: "Stop and remove container",
			Action:      client.RmCommand,
		},
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "start crane server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Value: "localhost:2475",
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
			Name:  "up",
			Usage: "Create and start container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "",
					Usage: "Load name (and optionally a tag) to be applied to the resulting",
				},
			},
			Action: client.UpCommand,
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
