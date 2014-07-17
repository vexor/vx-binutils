package main

import (
  "testing"
)

func TestNewLocalDir (t *testing.T) {

  dir, err := NewLocalDir("fixtures/local_files")

  if err != nil {
    t.Errorf("fail to open directory: %v", err)
  }

  if dir.Path != "fixtures/local_files" {
    t.Errorf("dir.Path must be %s, got %s", "", dir.Path)
  }

  if dir.Count != 2 {
    t.Errorf("dir.Count must be %d, got %d", 2, dir.Count)
  }

  if dir.Size != 28 {
    t.Errorf("dir.Size must be %d, got %d", 28, dir.Size)
  }

  if len(dir.Files) != 2 {
    t.Errorf("len(dir.Files) must be %d, got %d", 2, len(dir.Files))
  }

  file_0_name := dir.Files[0].Name
  if file_0_name != "directory/file.txt" {
    t.Errorf("dir.Files[0].Name must be %s, got %s", "directory/file.txt", file_0_name)
  }

  file_0_path := dir.Files[0].Path
  if file_0_path != "fixtures/local_files/directory/file.txt" {
    t.Errorf("dir.Files[0].Path must be %s, got %s", "fixtures/local_files/directory/file.txt", file_0_path)
  }

  file_0_size := dir.Files[0].Size
  if file_0_size != 19 {
    t.Errorf("dir.Files[0].Size must be %d, got %d", 19, file_0_size)
  }

  file_0_md5 := dir.Files[0].Md5
  if file_0_md5 != "3c75b2e00608caad58e47992854081f0" {
    t.Errorf("dir.Files[0].Md5 must be %s, got %s", "3c75b2e00608caad58e47992854081f0", file_0_md5)
  }

  if file_0_md5 != dir.Md5[file_0_name] {
    t.Errorf("dir.Md5[%s] must equal to %s, got %s", file_0_name, file_0_md5, dir.Md5[file_0_name])
  }

}

func TestNewLocalDirWithMissingDirectory (t *testing.T) {
  dir, err := NewLocalDir("fixtures/missing")

  if err == nil {
    t.Errorf("Error must be, got LocalDir %v", dir)
  }
}
