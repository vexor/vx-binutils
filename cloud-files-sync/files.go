package main

import (
  "os"
  "fmt"
  "path/filepath"
  "github.com/cheggaaa/pb"
  "crypto/md5"
  "io"
)

type File struct {
  Name string
  Size int64
  md5  string
}

type Files struct {
  count    int64
  size     int64
  files    []File
  md5      map[string]string
}

func (files *Files) CreateBar() *pb.ProgressBar {
  bar := pb.StartNew(int(files.size))
  bar.SetUnits(pb.U_BYTES)
  return bar
}

func (files *Files) visit(path string, f os.FileInfo, err error) error {

  if err != nil {
    return err
  }

  if !f.IsDir() {
    fio, err := os.Open(path)
    if err != nil {
      fmt.Printf("Fail to open file \"%s\"\n, message: \"%s\"\n", path, err)
      return err
    }

    md5 := md5.New()
    io.Copy(md5, fio)
    md5Value := fmt.Sprintf("%x", md5.Sum(nil))
    fio.Close()

    files.count += 1
    files.size  += f.Size()
    files.files = append(files.files, File{ Name: path, Size: f.Size(), md5: md5Value })
    files.md5[path] = md5Value
  }

  return nil
}

func NewFiles(dir string) (*Files, error) {
  files := &Files{
    md5: make(map[string]string),
  }
  if err := filepath.Walk(dir, files.visit) ; err != nil {
    fmt.Printf("Error occured during find files in \"%s\", original message \"%s\"\n", dir, err)
    return nil, err
  }

  fmt.Printf("Found %s in %d files\n", pb.FormatBytes(files.size), files.count)
  return files, nil
}
