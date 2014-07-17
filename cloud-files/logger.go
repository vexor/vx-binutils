package main

import (
  "github.com/wsxiaoys/terminal"
  "fmt"
  "os"
)

func LogInfo(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  terminal.Stdout.
    Colorf("@{y} --> %s", text).Nl().
    Reset()
}

func LogNl() {
  fmt.Printf("\n")
}

func LogInfoR(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  terminal.Stdout.
    ClearLine().
    Colorf("\r@{y} --> %s", text).
    Reset()
}

func LogWarnR(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  terminal.Stdout.
    ClearLine().
    Colorf("\r@{g} ==> %s", text).
    Reset()
}

func LogWarn(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  terminal.Stdout.
    Colorf("@{g} ==> %s", text).Nl().
    Reset()
}

func LogError(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  terminal.Stdout.
    Colorf("@{r}[ERROR] %s", text).Nl().
    Reset()
}

func LogFatal(err error) {
  if err != nil {
    LogNl()
    LogError(err.Error())
    os.Exit(1)
  }
}

func LogNotice(text string, a ...interface{}) {
  if len(a) > 0 {
    text = fmt.Sprintf(text, a...)
  }

  fmt.Printf("     %s\n", text)
}
