package main

import (
"bytes"
"fmt"
"image"
_ "image/png"
"os"
)

func main() {
if len(os.Args) < 2 {
fmt.Fprintln(os.Stderr, "usage: skininfo <png>")
os.Exit(1)
}
data, err := os.ReadFile(os.Args[1])
if err != nil {
fmt.Fprintf(os.Stderr, "read error: %v\n", err)
os.Exit(1)
}
if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
fmt.Fprintln(os.Stderr, "not a valid PNG (bad magic)")
os.Exit(1)
}
cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
if err != nil {
fmt.Fprintf(os.Stderr, "decode error: %v\n", err)
os.Exit(1)
}
fmt.Printf("%dx%d\n", cfg.Width, cfg.Height)
}
