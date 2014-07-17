package main

import (
  "os"
  "path/filepath"
  "fmt"
  "crypto/md5"
  "io"
)

type LocalFile struct {
  Name string
  Path string
  Size int64
  Md5  string
}

type LocalDir struct {
  Path     string
  Count    int64
  Size     int64
  Files    []LocalFile
  Md5      map[string]string
}

func (file LocalFile) OpenForRead() (*os.File, error) {
  io, err := os.Open(file.Path)
  if err != nil {
    return nil, err
  }

  return io, nil
}

func (dir *LocalDir) CreateFile(name string) (io.Writer, error) {
  fullPath := filepath.Join(dir.Path, name)
  dirName  := filepath.Dir(fullPath)

  if err := os.MkdirAll(dirName, 0755) ; err != nil {
    return nil, err
  }

  writer, err := os.Create(fullPath)
  if err != nil {
    return nil, err
  }

  return writer, nil
}

func (dir *LocalDir) visit(path string, f os.FileInfo, err error) error {


  if err != nil {
    return err
  }

  if !f.IsDir() {
    fio, err := os.Open(path)
    if err != nil {
      return err
    }

    md5 := md5.New()

    bytes, err := io.Copy(md5, fio)
    if err != nil && err != io.EOF {
      return err
    }

    md5Value := fmt.Sprintf("%x", md5.Sum(nil))
    err = fio.Close()
    if err != nil {
      return err
    }

    relPath, err := filepath.Rel(dir.Path, path)
    if err != nil {
      return err
    }

    dir.Count += 1
    dir.Size  += bytes
    dir.Files = append(dir.Files, LocalFile{ Name: relPath, Path: path, Size: bytes, Md5: md5Value })
    dir.Md5[relPath] = md5Value
  }

  return nil
}

func NewLocalDir(dirName string) (*LocalDir, error) {

  dirName = filepath.Clean(dirName)

  stat, err := os.Stat(dirName)
  if err != nil {
    return nil, err
  }

  dir := &LocalDir{
    Md5:  make(map[string]string),
  }

  if stat.IsDir() {
    dir.Path = dirName
    if err := filepath.Walk(dir.Path, dir.visit) ; err != nil {
      return nil, err
    }
  } else {
    dir.Path = filepath.Dir(dirName)
    if err := dir.visit(dirName, stat, nil) ; err != nil {
      return nil, err
    }
  }

  LogWarn("Found %d files in %q, total size %s", dir.Count, dir.Path, FormatBytes(dir.Size))
  return dir, nil
}

func CreateLocalDir(dirName string) (*LocalDir, error) {

  if err := os.MkdirAll(dirName, 0755) ; err != nil {
    return nil, err
  }

  dir := &LocalDir{
    Path: dirName,
    Md5:  make(map[string]string),
  }

  return dir, nil
}
