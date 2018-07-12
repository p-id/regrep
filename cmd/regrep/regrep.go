package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sort"

	"github.com/p-id/regrep/internal/index"
	"github.com/p-id/regrep/internal/regexp"
)

var usageMessage = `usage: regrep regexp [target-file-search] [-o target-result-output|stdout]

Regrep behaves like grep over files/target directory, searching for regexp,
an RE2 (nearly PCRE) regular expression.

The -c, -h, -i, -l, and -n flags are as in grep, although note that as per Go's
flag parsing convention, they cannot be combined: the option pair -i -n
cannot be abbreviated to -in.

The -o flag is optional and can be used to place output is a file instead to standard-out.

Regrep relies on the existence of an up-to-date index created ahead of time.
It builds or rebuilds the index if not present.

Regrep uses the index stored in $REGREPINDEX or, if that variable is unset or
empty, $HOME/.regrepindex.
`

var stdinFileName = `/dev/console/stdin`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

var (
	verboseFlag = flag.Bool("verbose", false, "print extra information")

	iFlag      = flag.Bool("i", false, "case-insensitive search")
	outputFile = flag.String("o", "", "write output to the specific file")
	cpuProfile = flag.String("cpuprofile", "", "write cpu profile to this file")

	matches bool
)

func Main() {
	g := regexp.Grep{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	g.AddFlags()

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 || len(args) > 2 {
		usage()
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			panic(err)
		}
		g.Stdout = file
		defer file.Close()
	}

	// Setup up the RegexpQuery
	pat := "(?m)" + args[0]
	if *iFlag {
		pat = "(?i)" + pat
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		log.Fatal(err)
	}
	g.Regexp = re

	q := index.RegexpQuery(re.Syntax)
	if *verboseFlag {
		log.Printf("query: %s\n", q)
	}

	// Step - 1 Always preprocess the files and build a trigram index

	// Translate paths to absolute paths so that we can
	// generate the file list in sorted order.
	if len(args) == 1 {
		args[0] = ""
	} else {
		args = args[1:]
		for i, arg := range args {
			a, err := filepath.Abs(arg)
			log.Printf("Including file %s: %s", arg, a)
			if err != nil {
				log.Printf("%s: %s", arg, err)
				args[i] = ""
				continue
			}
			args[i] = a
		}
		sort.Strings(args)
	}

	ixFile := index.File()
	// Always clear older index
	// We can always are additional logic to include pre-build index
	os.Remove(ixFile)

	wx := index.Create(ixFile)
	wx.Verbose = *verboseFlag

	var stdinBufferedReader io.Reader

	log.Printf("starting - index buildup")
	if len(args) == 1 && args[0] == "" {
		stdinReader := bufio.NewReader(os.Stdin)
		byteBuffer, err := ioutil.ReadAll(stdinReader)
		if err != nil {
			panic(err)
		}
		byteBufferReader := bytes.NewReader(byteBuffer)
		wx.Add(stdinFileName, byteBufferReader)
		byteBufferReader.Reset(byteBuffer)
		stdinBufferedReader = byteBufferReader
	} else {
		wx.AddPaths(args)
		for _, arg := range args {
			log.Printf("index %s", arg)

			filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if _, elem := filepath.Split(path); elem != "" {
					// Skip various temporary or "hidden" files or directories.
					if elem[0] == '.' || elem[0] == '#' || elem[0] == '~' || elem[len(elem)-1] == '~' {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
				}
				if err != nil {
					log.Printf("%s: %s", path, err)
					return nil
				}
				if info != nil && info.Mode()&os.ModeType == 0 {
					wx.AddFile(path)
				}
				return nil
			})
		}
	}
	log.Printf("flush index")
	wx.Flush()
	log.Printf("done - index buildup")

	var post []uint32
	rx := index.Open(ixFile)
	rx.Verbose = *verboseFlag
	post = rx.PostingQuery(q)
	if *verboseFlag {
		log.Printf("post query identified %d possible files\n", len(post))
	}

	for _, fileid := range post {
		if stdinBufferedReader != nil {
			g.Reader(stdinBufferedReader, stdinFileName)
		} else {
			name := rx.Name(fileid)
			g.File(name)
		}
	}

	matches = g.Match
}

func main() {
	Main()
	if !matches {
		os.Exit(1)
	}
	os.Exit(0)
}
