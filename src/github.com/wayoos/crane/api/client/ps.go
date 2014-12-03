package client

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	"github.com/wayoos/crane/api/domain"
	"os"
	"text/tabwriter"
)

func PsCommand(c *cli.Context) {

	host := c.GlobalString("host")
	result := []domain.LoadData{}
	resp, err := napping.Get(host+"/ps", nil, &result, nil)
	if err != nil {
		panic(err)
	}
	if resp.Status() == 200 {
		w := new(tabwriter.Writer)

		// Format in tab-separated columns with a tab stop of 8.
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "DOCKLOAD ID\tSTATUS\tNAME\tTAG")
		for _, loadData := range result {
			fmt.Fprintln(w, loadData.ID+"\t"+""+"\t"+loadData.Name+"\t"+loadData.Tag)
		}
		w.Flush()

	}

}
