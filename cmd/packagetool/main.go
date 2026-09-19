package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Ambloplites/packagetool"
)

func main() {
	var (
		out        string
		list       bool
		useDecrypt bool
	)

	flag.StringVar(&out, "o", "./", "Path to the output directory")
	flag.BoolVar(&list, "l", false, "List archive contents")
	flag.BoolVar(&useDecrypt, "d", false, "Enable MAR decryption")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILENAME\nList of available options:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(filename)
	if err != nil {
		fatal(err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)

	pr, err := packagetool.DetectArchiveType(rd)
	if err != nil {
		fatal("failed to determine archive type:", err)
	}
	fmt.Println("archive type:", pr.Name())
	if mar, ok := pr.(*packagetool.MarReader); ok {
		mar.UseDecryption = useDecrypt
	}

	var callback func(packagetool.File) error
	if list {
		callback = listArchive
	} else {
		dumper := &packagetool.Dumper{
			Dest: out,
		}
		callback = dumper.DumpArchive
	}

	start := time.Now()
	if err := pr.Read(rd, callback); err != nil {
		fatal(err)
	}
	fmt.Println("time elapsed:", time.Since(start))
}

func listArchive(f packagetool.File) error {
	fmt.Println(f.Filename)
	f.Skip()
	return nil
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
