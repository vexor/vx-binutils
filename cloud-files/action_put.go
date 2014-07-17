package main

import (
  "os"
  "github.com/spf13/cobra"
  "time"
)

type ActionPutReq struct {
  c  *Container
  f  LocalFile
}

type ActionPutReply struct {
  o   *ContainerObject
  c   *Container
  err error
}

type ActionPutDelReq struct {
  c   *Container
  o   string
}

type ActionPutDestroyList struct {
  c     *Container
  names []string
  total int
}

type ActionPut struct {
  chIn       chan ActionPutReq
  chOut      chan ActionPutReply
  remove     bool
  startAt    time.Time
  dir        *LocalDir
  containers []*Container
  nthreads   int
  progress   *Progress
  totalFiles int
  prefix     string
}

func (put ActionPut) Close() {
  close(put.chIn)
  close(put.chOut)
}

func (put ActionPut) BeginUpload() {
  go func() {
    for c := range put.containers {
      for f := range put.dir.Files {
        req := ActionPutReq{
          c: put.containers[c],
          f: put.dir.Files[f],
        }
        put.chIn <- req
      }
    }
  }()
}

func (put ActionPut) Upload () {
  for i := 0 ; i < put.nthreads ; i++ {
    go func() {
      for req := range(put.chIn) {
        obj, err := req.c.UploadFile(req.f, put.progress, put.prefix)
        put.chOut <- ActionPutReply{ c: req.c, o: obj, err: err }
      }
    }()
  }
}

func (put ActionPut) WaitUpload() {

  total := put.totalFiles

  for reply := range put.chOut {

    total--

    if reply.err != nil {
      LogFatal(reply.err)
    }

    if total == 0 {
      break
    }
  }
}

func (put ActionPut) RemoveObjectsPool(in chan ActionPutDelReq, out chan error) {
  for i := 0 ; i < put.nthreads ; i ++ {
    go func() {
      for req := range(in) {
        err := req.c.conn.ObjectDelete(req.c.Name, req.o)
        out <- err
      }
    }()
  }
}

func (put ActionPut) RemoveObjects() {
  in  := make(chan ActionPutDelReq)
  out := make(chan error)

  defer close(in)
  defer close(out)

  put.RemoveObjectsPool(in, out)

  for _, c := range put.containers {
    objectNames, err := c.ListObjects()
    if err != nil {
      LogError("[%s:%s] List objects errors: %s", c.Region, c.Name, err.Error())
      os.Exit(1)
    }

    toDestroy := []string{}
    for _, o := range objectNames {
      name := o

      if prefix != "" {
        if len(o) < len(prefix) {
          name = ""
        } else {
          name = o[len(prefix):len(o)]
        }
      }

      if put.dir.Md5[name] == "" {
        toDestroy = append(toDestroy, o)
      }
    }

    if len(toDestroy) > 0 {
      LogNl()
      go func(){
        for _, o := range toDestroy {
          in <- ActionPutDelReq{ c: c, o: o }
        }
      }()

      total := len(toDestroy)

      for err := range out {
        LogFatal(err)
        total--

        LogInfoR("[%s:%s] delete %d of %d", c.Region, c.Name, len(toDestroy) - total, len(toDestroy))
        if total == 0 {
          break
        }
      }
    }
  }
}

func ActionPutHandler(cmd *cobra.Command, args []string) {

  directoryName := cmd.Flag("directory").Value.String()
  removeFiles   := cmd.Flag("delete").Value.String()
  prefix        := cmd.Flag("prefix").Value.String()

  if len(args) == 0 {
    LogNl()
    LogError("Container name must be set")
    os.Exit(1)
  }

  dir, err := NewLocalDir(directoryName)
  LogFatal(err)

  containerName   := args[0]
  containers, err := FindContainers(containerName)
  LogFatal(err)

  totalBytes := int64(len(containers)) * dir.Size

  put := ActionPut{
    nthreads:   20,
    chIn:       make(chan ActionPutReq,   20),
    chOut:      make(chan ActionPutReply, 20),
    startAt:    time.Now(),
    dir:        dir,
    containers: containers,
    progress:   NewProgress(totalBytes),
    totalFiles: len(containers) * int(dir.Count),
    prefix:     prefix,
  }

  if removeFiles == "true" {
    put.remove = true
  }

  defer put.Close()

  if dir.Count > 0 {
    put.progress.Begin()
    put.BeginUpload()
    put.Upload()
    put.WaitUpload()
  }

  if put.remove {
    put.RemoveObjects()
  }

  LogNl()
  LogWarn("DONE.")
}
