package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	flagDownloadLatestAssets = flag.Bool("D", false, "Download and read the latest assets from GitHub")
	flagInlineTemplate       = flag.String("inline", "", "Read the template of the manifest JSON from the argument")
	flagStdinTemplate        = flag.Bool("stdin", false, "Read the template of the manifest JSON from the standard input")
	flagUserAndRepo          = flag.String("g", "", "Specify GitHub's \"USER/REPOSITORY\"")
	flagAnyCPU               = flag.Bool("anycpu", false, "Do not use \"architecture\" of the manifest")
	flagExtractDir           = flag.Bool("p", false, "Specify the parent directory of *.exe into \"extract_dir\" and the basename into \"bin\"")
	flag32                   = flag.String("32", "386,486,586,686,32bit,win32", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 32bit")
	flag64                   = flag.String("64", "amd64,64bit,win64,x86_64,x64", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 64bit")
	flagLicense              = flag.String("license", "", "Set the value of \"license\" of the manifest")
	flagDescription          = flag.String("description", "", "Set the value of \"description\" of the manifest")
	flagDownloadTo           = flag.String("downloadto", "", "Do not remove the downloaded zip files and save them onto the specified directory")
	flagBinPattern           = flag.String("binpattern", "*.exe", "The pattern for executables(separated with comma)")
	flagIgnoreWords          = flag.String("ignore", "linux,macos,freebsd,netbsd,darwin,plan9", "ignore the zipfile whose name contains these words")
	flagNoAutoUpdate         = flag.Bool("noautoupdate", false, "disable autoupdate")
)

var version string

func mains(args []string) error {
	if len(args) > 0 || *flagDownloadLatestAssets || *flagDownloadTo != "" {
		return tryGithub(args)
	} else {
		flag.PrintDefaults()
		return nil
	}
}

func main() {
	flag.Parse()

	fmt.Fprintf(os.Stderr, "%s %s for %s/%s by %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())

	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
