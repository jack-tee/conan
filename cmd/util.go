package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func GetPersistentFlags(cmd *cobra.Command) (string, string) {
	host, err := cmd.Root().PersistentFlags().GetString("host")
	if err != nil {
		log.Fatalln(err)
	}

	port, err := cmd.Root().PersistentFlags().GetString("port")
	if err != nil {
		log.Fatalln(err)
	}

	return host, port
}
