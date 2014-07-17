package main

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
  "io/ioutil"
  "log"
)

var (
  directoryName   string
  containerName   string
  numberOfThreads uint
  action          string
  destroy         bool
  prefix          string
)

func validateConatinerName() {
  if containerName == "" {
    fmt.Printf("container name must be set")
    os.Exit(1)
  }
}

func main() {

  log.SetOutput(ioutil.Discard)

  var cmdPut = &cobra.Command{
    Use:   "put container" ,
    Short: "Upload local directory or file to remote CloudFile container",
    Run:   ActionPutHandler,
  }

  var cmdGet = &cobra.Command{
    Use:   "get container" ,
    Short: "Get object(s) from remote CloudFile container",
    Run:   ActionGetHandler,
  }

  cmdPut.Flags().BoolVarP(&destroy, "delete", "d", false, "delete extraneous files from container")
  cmdPut.Flags().StringVarP(&directoryName, "directory", "s", ".", "source directory or file to upload")
  cmdPut.Flags().StringVarP(&prefix, "prefix", "p", "", "add prefix to each uploaded object")

  cmdGet.Flags().StringVarP(&directoryName, "directory", "s", ".", "directory name to download")

  var rootCmd = &cobra.Command{Use: os.Args[0]}
  rootCmd.AddCommand(cmdGet, cmdPut)
  rootCmd.Execute()
}
