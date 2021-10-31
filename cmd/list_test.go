package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
)

func Test_ListCommand(t *testing.T) {
	cmd := listCmd
	b := bytes.NewBufferString("")
	log.SetOutput(b)
	cmd.SetOut(b)
	//cmd.SetArgs([]string{"list"})
	cmd.Flags().Set("host", "woop")
	cmd.Root().Flags().Set("host", "blah")
	f, _ := cmd.PersistentFlags().GetString("host")
	fmt.Println("flag", f)

	cmd.Execute()
	fmt.Println("short", cmd.Short)
	out, _ := ioutil.ReadAll(b)
	fmt.Println("hi", string(out))
}
