/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var host string
var port string
var debug bool
var templates *template.Template
var templatesPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conan",
	Short: "A simple connector manager for Kafka Connect",
	Long:  `A simple connector manager for Kafka Connect.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	//fmt.Println("Running root.go init")
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.conan.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "the Kafka Connect rest api host")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8083", "the Kafka Connect rest api port")
	rootCmd.PersistentFlags().StringVar(&templatesPath, "templatesPath", "templates/*.tmpl", "path to output templates")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}

func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		log.Info("Debug logs enabled")
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{})

	templates = template.Must(template.New("").Parse(defaultTemplates))
	log.Debug("Loading templates using path ", templatesPath)
	templates.ParseGlob(templatesPath)

	log.Debug("Templates loaded.", templates.DefinedTemplates())
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Find home directory.
// 		home, err := os.UserHomeDir()
// 		cobra.CheckErr(err)

// 		// Search config in home directory with name ".conan" (without extension).
// 		viper.AddConfigPath(home)
// 		viper.SetConfigType("yaml")
// 		viper.SetConfigName(".conan")
// 	}

// 	viper.AutomaticEnv() // read in environment variables that match

// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
// 	}
// }
