package main

import (
  "github.com/ncw/swift"
  "os"
  "errors"
  "strings"
  "sync"
  "fmt"
  "io"
)

type Container struct {
  Region    string
  Name      string
  conn      *swift.Connection
  nthreads  uint
}

type ContainerObject struct {
  writer    *swift.ObjectCreateFile
  reader    *swift.ObjectOpenFile
  Md5       string
  Size      int64
  Name      string
}

type ContainerSpec struct {
  Name      string
  Regions   []string
}

func (c *Container) Valid() error {

  if err := c.conn.Authenticate() ; err != nil {
    return err
  }

  _, _, err := c.conn.Container(c.Name)
  if err != nil {
    return err
  }

  return nil
}

func (c *Container) FindObject(name string) (*ContainerObject, error) {
  a, h, err := c.conn.Object(c.Name, name)
  if err != nil {
    return nil, err
  }

  obj := &ContainerObject{
    Md5:     h["Origin"],
    Size:    a.Bytes,
    Name:    name,
  }

  return obj, nil
}

func (c *Container) CreateObject(file LocalFile, prefix string) (*ContainerObject, error) {
  name := prefix + file.Name
  h    := swift.Headers{}
  h["Origin"] = file.Md5 // CloudFiles does not support custom headers, using allowed
  origin, err := c.conn.ObjectCreate(c.Name, name, true, file.Md5, "", h)
  if err != nil {
    return nil, err
  }

  obj := &ContainerObject{
    writer:  origin,
    Md5:     file.Md5,
    Size:    file.Size,
    Name:    name,
  }

  return obj, nil
}

func (c *Container) OpenObject(name string) (*ContainerObject, error) {
  reader, h, err := c.conn.ObjectOpen(c.Name, name, false, swift.Headers{})

  if err != nil {
    return nil, err
  }

  bytes, err := reader.Length()
  if err != nil {
    return nil, err
  }

  obj := &ContainerObject{
    reader:  reader,
    Md5:     h["Origin"],
    Size:    bytes,
    Name:    name,
  }

  return obj, err
}

func (c *Container) UploadFile(file LocalFile, p *Progress, prefix string) (*ContainerObject, error) {

  name := prefix + file.Name
  if obj, err := c.FindObject(name) ; err == nil && obj.Md5 == file.Md5 {
    p.Add(obj.Size)
    return obj, nil
  }

  rio, err := file.OpenForRead()
  if err != nil {
    return nil, err
  }

  defer rio.Close()

  obj, err := c.CreateObject(file, prefix)
  if err != nil {
    return nil, err
  }

  defer func() {
    if err := obj.writer.Close() ; err != nil && err != io.EOF {
      LogNl()
      LogError("[%s:%s] %s", c.Region, c.Name, file.Name)
      LogError(err.Error())
      os.Exit(1)
    }
  }()

  writer := io.MultiWriter(obj.writer, p)

  bytes, err := io.Copy(writer, rio)
  if err != nil {
    return nil, err
  }

  obj.Size = bytes

  return obj, nil
}

func (c *Container) ListObjects() ([]string, error) {
  opts := swift.ObjectsOpts{}
  objects, err := c.conn.ObjectNamesAll(c.Name, &opts)
  if err != nil {
    return objects, err
  }

  return objects, nil
}

func ParseContainerSpec(spec string) *ContainerSpec {
  regions := []string{ "IAD", "DFW", "ORD", "HKG", "SYD", "LON" }

  specList := strings.Split(spec, ":")

  if len(specList) >= 2 {
    return &ContainerSpec{
      Name: specList[1],
      Regions: []string{ strings.ToUpper(specList[0]) },
    }
  }

  if len(specList) == 1 {
    return &ContainerSpec{
      Name:    specList[0],
      Regions: regions,
    }
  }

  return &ContainerSpec{}
}

func FindContainers(containerSpec string) ([]*Container, error) {

  var containers    []*Container

  userName      := os.Getenv("SDK_USERNAME")
  apiKey        := os.Getenv("SDK_API_KEY")

  if userName == "" || apiKey == "" {
    return containers, errors.New("Missing SDK_USERNAME or SDK_API_KEY environment variables")
  }

  spec := ParseContainerSpec(containerSpec)

  LogWarn("Using regions: %q", spec.Regions)

  wg := new(sync.WaitGroup)
  m  := new(sync.Mutex)

  for _, region := range(spec.Regions) {

    //authUrl := "https://auth.api.rackspacecloud.com/v1.0"
    authUrl := "https://identity.api.rackspacecloud.com/v2.0"
    if region == "LON" {
      authUrl = "https://lon.identity.api.rackspacecloud.com/v2.0"
    }

    conn := &swift.Connection{
      UserName: userName,
      ApiKey:   apiKey,
      Region:   region,
      AuthUrl:  authUrl,
    }

    container := &Container{
      Region:   region,
      Name:     spec.Name,
      conn:     conn,
      nthreads: 20,
    }

    wg.Add(1)
    go func() {
      err := container.Valid()
      if err == nil {
        m.Lock()
        containers = append(containers, container)
        m.Unlock()
      } else {
        LogNotice("%v", err)
      }
      wg.Done()
    }()
  }

  wg.Wait()

  if len(containers) > 0 {
    LogWarn("Found container %q in %d regions", spec.Name, len(containers))
    return containers, nil
  } else {
    return containers, errors.New(fmt.Sprintf("Cannot found container %q in %q regions", spec.Name, spec.Regions))
  }
}
