package main

import (
  "github.com/spf13/cobra"
  "os"
  "io"
)

type ActionGet struct {
  container    *Container
  directory    string
  totalObjects int64
  totalBytes   int64
  objectNames  []string
  dir          *LocalDir
}

type ActionGetFindRes struct {
  o  *ContainerObject
  err error
}

func (g *ActionGet) Download() {

  in  := make(chan string)
  out := make(chan error)

  p := NewProgress(g.totalBytes)
  p.Begin()

  for i := 0 ; i < 2 ; i++ {
    go func() {
      for name := range(in) {
        obj, err := g.container.OpenObject(name)
        if err != nil {
          out <- err
          break
        }

        file, err := g.dir.CreateFile(obj.Name)
        if err != nil {
          out <- err
          break
        }

        writer := io.MultiWriter(file, p)

        _, err = io.Copy(writer, obj.reader)
        if err != nil {
          out <- err
          break
        }

        out <- nil
      }
    }()
  }

  go func() {
    for _, name := range g.objectNames {
      in <- name
    }
  }()

  total := g.totalObjects

  for err := range out {
    total--

    if err != nil {
      LogFatal(err)
    }

    if total == 0 {
      break
    }
  }
}

func (g *ActionGet) CollectObjects () {
  in  := make(chan string, 20)
  out := make(chan ActionGetFindRes, 20)

  defer close(in)
  defer close(out)

  for i := 0 ; i < 5 ; i++ {
    go func() {
      for name := range in {
        obj, err := g.container.FindObject(name)
        out <- ActionGetFindRes{ o: obj, err: err }
      }
    }()
  }

  go func() {
    for _, name := range g.objectNames {
      in <- name
    }
  }()

  total := g.totalObjects

  for res := range(out) {
    total--

    if res.err != nil {
      LogFatal(res.err)
    }

    g.totalBytes += res.o.Size
    LogInfoR("Collect %d of %d", g.totalObjects - total, g.totalObjects)
    if total == 0 {
      break
    }
  }

  LogNl()
}

func ActionGetHandler(cmd *cobra.Command, args []string) {
  directoryName := cmd.Flag("directory").Value.String()

  if len(args) == 0 {
    LogNl()
    LogError("Container name must be set")
    os.Exit(1)
  }

  dir, err := CreateLocalDir(directoryName)
  LogFatal(err)

  containerName   := args[0]
  containers, err := FindContainers(containerName)
  LogFatal(err)

  container := containers[0]
  LogWarn("Using %q region", container.Region)

  objectNames, err := container.ListObjects()
  LogFatal(err)

  get := &ActionGet{
    directory:    dir.Path,
    container:    container,
    totalObjects: int64(len(objectNames)),
    objectNames:  objectNames,
    dir:          dir,
  }

  get.CollectObjects()
  get.Download()
  LogNl()
  LogWarn("DONE.")
}
