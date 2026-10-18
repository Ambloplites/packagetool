package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/Ambloplites/fsdump"
	"github.com/Ambloplites/packagetool"
)

func main() {
	var (
		out        string
		list       bool
		workers    int
		useDecrypt bool
	)

	flag.StringVar(&out, "o", "./", "Path to the output directory")
	flag.BoolVar(&list, "l", false, "List archive contents")
	flag.IntVar(&workers, "w", 0, "Number of workers. Specify a value less than one, and the number of logical CPUs available to the process will be used")
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

	if workers < 0 {
		workers = runtime.NumCPU()
	}

	f, err := os.Open(filename)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	prd, err := packagetool.NewPackageReader(f, 0, packagetool.PackageReaderSettings{DecryptMAR: useDecrypt})
	if err != nil {
		fatal(err)
	}

	if list {
		for file, err := range prd.Files() {
			if err != nil {
				fatal("failed to read file entry:", err)
			}
			fmt.Println(file.CleanPath())
		}
	} else {
		pkg, err := packagetool.NewPackage(prd)
		if err != nil {
			fatal("failed to load package:", err)
		}

		copied := 0
		dumper := fsdump.Dumper{
			Src:        &fsdump.ChannelFileSource{Chan: pkg.Dump()},
			Dest:       out,
			NumWorkers: min(workers, pkg.FileCount()),
			ProgressCallback: func(status fsdump.DumpStatus) {
				if status.Err != nil {
					fmt.Fprintf(os.Stderr, "failed to copy file %q: %v\n", status.Path, status.Err)
				} else {
					copied++
					fmt.Printf("[%d/%d] %s\n", copied, pkg.FileCount(), status.Path)
				}
			},
		}
		dumper.Run()

		if copied < pkg.FileCount() {
			fatal("finished with", pkg.FileCount()-copied, "error(s)")
		}
	}
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
