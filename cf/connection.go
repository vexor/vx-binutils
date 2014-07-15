package main

import (
  "io"
  "os"
  "fmt"
  "github.com/ncw/swift"
  "github.com/cheggaaa/pb"
  "sync"
  "path/filepath"
  "crypto/md5"
  "errors"
)

type Object struct {
  swift.Object
  md5   string
}

type Connection struct {
  conn      *swift.Connection
  container string
  nthreads  uint
  objects   []string
}

type SyncTask struct {
  file File
  pb   *pb.ProgressBar
}

type SyncReply struct {
  file File
  err  error
}

type FetchTask struct {
  name string
  pb   *pb.ProgressBar
}

type FetchReply struct {
  name   string
  err    error
}

func (c *Connection) FindContainer(name string) error {
  _, _, err := c.conn.Container(name)

  if err != nil {
    fmt.Printf("Container \"%s\" does not exists, message \"%s\"\n", name, err)
    return err
  }

  fmt.Printf("Found container \"%s\"\n", name)
  return nil
}

func (c *Connection) CreateContainer(name string) error {
  if err := c.conn.ContainerCreate(name, swift.Headers{}) ; err != nil {
    fmt.Printf("Fail to create new container \"%s\", message: \"\"\n", name, err)
    return err
  } else {
    fmt.Printf("Successfuly create container \"%s\"\n", name)
  }
  return nil
}

func (c *Connection) FindOrCreateContainer(name string) error {
  err := c.FindContainer(name)

  if err == nil {
    return nil
  }

  err = c.CreateContainer(name)

  if err == nil {
    return nil
  }

  err = c.FindContainer(name)
  return err
}

func (c *Connection) ListObjects() ([]string, error) {
  opts := swift.ObjectsOpts{}
  objects, err := c.conn.ObjectNamesAll(c.container, &opts)

  if err != nil {
    fmt.Printf("Error listing objects in \"%s\", message: \"%s\"\n", c.container, err)
    return objects, err
  }

  return objects, nil
}

func (c *Connection) FindObjectMd5(name string) (string, error) {
  _, h, err := c.conn.Object(c.container, name)

  if err != nil {
    return "", err
  }

  return h["Origin"], err
}

func (c *Connection) FindObject(name string) (Object, error) {
  o, h, err := c.conn.Object(c.container, name)
  if err != nil {
    fmt.Printf("Fail to get object \"%s\" from \"%s\", message: \"%s\"\n", name, c.container, err)
    return Object{}, err
  }

  object := Object{Object: o, md5: h["Origin"]}

  return object, nil
}

func (c *Connection) CreateObject(name string, md5 string) (*swift.ObjectCreateFile, error) {
  h := swift.Headers{}
  h["Origin"] = md5 // CloudFiles does not support custom headers, using allowed
  return c.conn.ObjectCreate(c.container, name, true, md5, "", h)
}

func (c *Connection) UploadFile(file File, pb *pb.ProgressBar) error {

  md5, err := c.FindObjectMd5(file.Name)
  if md5 == file.md5 {
    pb.Add(int(file.Size))
    return nil
  }

  rio, err := os.Open(file.Name)
  if err != nil {
    fmt.Printf("fail to open file \"%s\", message: \"%s\"\n", file.Name, err)
    return err
  }

  defer func() {
    if err := rio.Close() ; err != nil {
      panic(err)
    }
  }()

  wio, err := c.CreateObject(file.Name, file.md5)
  if err != nil {
    fmt.Printf("fail to create remote object \"%s\", message: \"%s\"\n", file.Name, err)
    return err
  }

  defer func() {
    if err := wio.Close() ; err != nil {
      panic(err)
    }
  }()

  writer := io.MultiWriter(wio, pb)

  _, err = io.Copy(writer, rio)
  if err != nil {
    fmt.Printf("Fail write to \"%s\", message \"%s\"\n", file.Name, err)
    return err
  }

  return nil
}

func (c *Connection) UploadWorker(tasks chan SyncTask, reply chan SyncReply, wg *sync.WaitGroup) {

  defer wg.Done()

  for task := range tasks {
    err := c.UploadFile(task.file, task.pb)
    reply <- SyncReply{ file: task.file, err: err }
  }
}

func (c *Connection) UploadFiles(files *Files, pb *pb.ProgressBar) error {
  wg := new(sync.WaitGroup)
  reply := make(chan SyncReply, c.nthreads)
  tasks := make(chan SyncTask, c.nthreads)

  defer close(reply)

  for i := uint(0) ; i < c.nthreads ; i++ {
    wg.Add(1)
    go c.UploadWorker(tasks, reply, wg)
  }

  go func() {
    for _, file := range files.files {
      task := SyncTask{
        file: file,
        pb:   pb,
      }
      tasks <- task
    }
    close(tasks)
  }()

  n := 0
  for r := range reply {
    n += 1
    if r.err != nil {
      fmt.Printf("\nFail to upload file %s, message: %s\n", r.file.Name, r.err)
      return r.err
    }
    if n == len(files.files) {
      break
    }
  }

  wg.Wait()

  return nil
}

func (c *Connection) Sync(dir string) error {
  files, err := NewFiles(dir)

  if err != nil {
    return err
  }

  bar := files.CreateBar()
  bar.Start()

  if err := c.UploadFiles(files, bar) ; err != nil {
    return err
  }
  bar.Finish()

  var toRemove []string
  objects, err := c.ListObjects()
  if err != nil {
    return err
  }

  for _, obj := range objects {
    if files.md5[obj] == "" {
      toRemove = append(toRemove, obj)
    }
  }

  if len(toRemove) > 0 {
    bar = pb.StartNew(len(toRemove))

    for _, obj := range toRemove {
      bar.Increment()
      if err := c.conn.ObjectDelete(c.container, obj) ; err != nil {
        fmt.Printf("Fail to remove \"%s\" from \"%s\", message: \"%s\"", obj, c.container, err)
        return err
      }
    }
    bar.Finish()
  }

  fmt.Printf("Done.\n")

  return nil
}

func (c *Connection) FetchObject(dirName string, objName string, pb *pb.ProgressBar) error {
  rio, h, err := c.conn.ObjectOpen(c.container, objName, false, swift.Headers{})
  if err != nil {
    fmt.Printf("Fail to open object \"%s\", message: \"%s\"\n", objName, err)
    return err
  }

  objPath := fmt.Sprintf("%s/%s", dirName, objName)
  objDir  := filepath.Dir(objPath)

  err = os.MkdirAll(objDir, 0755)
  if err != nil {
    fmt.Printf("Fail to create directory \"%s\", message: \"%s\"\n", objDir, err)
    return err
  }

  wio, err := os.Create(objPath)
  if err != nil {
    fmt.Printf("Fail to open file \"%s\", message: \"%s\"\n", objPath, err)
    return err
  }

  md5 := md5.New()

  writer := io.MultiWriter(wio, pb, md5)
  io.Copy(writer, rio)

  if err := wio.Close() ; err != nil {
    panic(err)
  }

  if err := rio.Close() ; err != nil {
    panic(err)
  }

  md5Value := fmt.Sprintf("%x", md5.Sum(nil))
  if  md5Value != h["Origin"] {
    err := errors.New(fmt.Sprintf("Checksum on \"%s\" fail expected \"%s\", got \"%s\"", objName, md5Value, h["Origin"]))
    fmt.Printf("%s\n", err)
    return err
  }

  return nil
}

func (c *Connection) FetchWorker(dirName string, tasks chan FetchTask, reply chan FetchReply, wg *sync.WaitGroup) {

  for task := range tasks {
    err := c.FetchObject(dirName, task.name, task.pb)
    reply <- FetchReply{ err: err, name: task.name }
  }
  wg.Done()

}

func (c *Connection) FetchContainer(dirName string) error {

  objectNames, err := c.ListObjects()
  if err != nil {
    return err
  }

  cont, _, err := c.conn.Container(c.container)
  if err != nil {
    fmt.Printf("Fail to find container \"%s\", message: \"%s\"\n", c.container, err)
    return err
  }

  wg    := new(sync.WaitGroup)
  tasks := make(chan FetchTask,  c.nthreads)
  reply := make(chan FetchReply, c.nthreads)

  defer close(reply)

  for i := uint(0) ; i < c.nthreads ; i++ {
    wg.Add(1)
    go c.FetchWorker(dirName, tasks, reply, wg)
  }

  bar := pb.StartNew(int(cont.Bytes))
  bar.SetUnits(pb.U_BYTES)

  go func() {
    for _, objName := range objectNames {
      task := FetchTask{
        name: objName,
        pb:   bar,
      }
      tasks <- task
    }
    close(tasks)
  }()

  n := 0
  for r := range reply {
    n += 1
    if r.err != nil {
      fmt.Printf("\nFail to download file %s, message: %s\n", r.name, r.err)
      return r.err
    }
    if n == len(objectNames) {
      break
    }
  }

  wg.Wait()

  bar.FinishPrint("Done.")

  return nil
}

func NewConnection(container string, nthreads uint) (*Connection, error) {

  url    := os.Getenv("SDK_AUTH_URL")
  region := os.Getenv("SDK_REGION")

  if url == "" {
    url = "https://auth.api.rackspacecloud.com/v1.0"
  }

  if region == "" {
    region = "IAD"
  }

  conn := swift.Connection{
    UserName: os.Getenv("SDK_USERNAME"),
    ApiKey:   os.Getenv("SDK_API_KEY"),
    Region:   region,
    AuthUrl:  url,
  }

  fmt.Printf(
    "Connecting using UserName: %s, Region: %s, AuthUrl: %s\n",
    conn.UserName,
    conn.Region,
    conn.AuthUrl,
  )

  if err := conn.Authenticate() ; err != nil {
    fmt.Printf("%s\n", err)
    return nil, err
  }

  fmt.Println("Successfuly connected")

  c := &Connection{
    conn:     &conn,
    nthreads: nthreads,
  }

  if err := c.FindOrCreateContainer(container) ; err != nil {
    return nil, err
  }

  c.container = container

  return c, nil
}
