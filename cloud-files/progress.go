package main

import (
  "io"
  "time"
  "fmt"
  "sync/atomic"
)

type Progress struct {
  io.Writer
  startAt    time.Time
  totalBytes int64
  curBytes   int64
}

func (p *Progress) String() string {
  s := fmt.Sprintf("%.2f %% | %s / sec", p.Percent(), FormatBytes(int64(p.Value())))
  return s
}

func (p *Progress) Percent() float64 {
  per := float64(p.curBytes) / float64(p.totalBytes) * 100.0
  return per
}

func (p *Progress) Value() float64 {
  tm := time.Now().Sub(p.startAt).Seconds()
  value := float64(p.curBytes) / tm
  return value
}

func (p *Progress) Write(b []byte) (int, error) {
  bytes := len(b)
  p.Add(int64(bytes))
  return bytes, nil
}

func (p *Progress) Add(n int64) {
  atomic.AddInt64(&p.curBytes, n)
}

func (p *Progress) Begin() {
  p.startAt = time.Now()

  go func() {
    Loop:
      for {
        select {
        case <- time.After(1 * time.Second):
          LogInfoR(p.String())
        }

        if p.totalBytes == p.curBytes {
          break Loop
        }
      }
    LogInfoR(p.String())
  }()
}

func NewProgress(totalBytes int64) *Progress {
  p := &Progress{
    totalBytes: totalBytes,
  }

  return p
}
