package main

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
)

var (
  directoryName   string
  containerName   string
  numberOfThreads uint
  action          string
)

func validateConatinerName() {
  if containerName == "" {
    fmt.Printf("container name must be set")
    os.Exit(1)
  }
}

func main() {

  var cmdSync = &cobra.Command{
    Use:   "sync -c container",
    Short: "Sync local directory to remote CloudFile container",
    Run: func(cmd *cobra.Command, args []string) {
      validateConatinerName()
      c, err := NewConnection(containerName, numberOfThreads)
      if err != nil {
        os.Exit(1)
      }

      err = c.Sync(directoryName)
      if err != nil {
        os.Exit(1)
      }
    },
  }

  var cmdFetch = &cobra.Command{
    Use:   "fetch -c container",
    Short: "Sync remote CloudFile container to local directory",
    Run: func(cmd *cobra.Command, args []string) {
      validateConatinerName()
      c, err := NewConnection(containerName, numberOfThreads)
      if err != nil {
        os.Exit(1)
      }

      err = c.FetchContainer(directoryName)
      if err != nil {
        os.Exit(1)
      }
    },
  }

  cmdSync.Flags().StringVarP(&directoryName, "directory", "d", ".", "directory name to upload")
  cmdSync.Flags().StringVarP(&containerName, "container", "c", "", "destination container name")
  cmdSync.Flags().UintVarP(&numberOfThreads, "threads", "t", 10, "number of concurent threads")

  cmdFetch.Flags().StringVarP(&directoryName, "directory", "d", ".", "directory name to download")
  cmdFetch.Flags().StringVarP(&containerName, "container", "c", "", "source container name")
  cmdFetch.Flags().UintVarP(&numberOfThreads, "threads", "t", 10, "number of concurent threads")

  var rootCmd = &cobra.Command{Use: os.Args[0]}
  rootCmd.AddCommand(cmdSync, cmdFetch)
  rootCmd.Execute()
}
