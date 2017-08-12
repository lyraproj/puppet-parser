package testutils

import (
  "bytes"
  "math"
)

// For test purposes only.
//
// Determines the maximum indent that can be stripped by looking at leading whitespace on all lines. Lines that
// consists entirely of whitespace are not included in the computation.
// Strips first line if it's empty, then strips the computed indent from all lines and returns the result.
//
func Unindent(str string) string {
  minIndent := computeIndent(str)
  if minIndent == 0 {
    return str
  }
  r := bytes.NewBufferString(str)
  b := bytes.NewBufferString("")
  first := true
  for {
    line, err := r.ReadString('\n')
    if first {
      first = false
      if line == "\n" {
        continue
      }
    }
    if len(line) > minIndent {
      b.WriteString(line[minIndent:])
    } else if err == nil {
      b.WriteByte('\n')
    } else {
      break
    }
  }
  return b.String()
}

func computeIndent(str string) int {
  minIndent := math.MaxInt64
  r := bytes.NewBufferString(str)
  for minIndent > 0 {
    line, err := r.ReadString('\n')
    ll := len(line)

    for wsCount := 0; wsCount < minIndent && wsCount < ll; wsCount++ {
      c := line[wsCount]
      if c != ' ' && c != '\t' {
        if c != '\n' {
          minIndent = wsCount
        }
        break
      }
    }
    if err != nil {
      break
    }
  }
  if minIndent == math.MaxInt64 {
    minIndent = 0
  }
  return minIndent
}

