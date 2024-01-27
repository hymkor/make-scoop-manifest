package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hymkor/make-scoop-manifest/internal/gitdir"
	"github.com/hymkor/make-scoop-manifest/internal/github"
)

var (
	flagDownloadLatestAssets = flag.Bool("D", false, "Download and read the latest assets from GitHub")
	flagInlineTemplate       = flag.String("inline", "", "Read the template of the manifest JSON from the argument")
	flagStdinTemplate        = flag.Bool("stdin", false, "Read the template of the manifest JSON from the standard input")
	flagUserAndRepo          = flag.String("g", "", "Specify GitHub's \"USER/REPOSITORY\"")
	flagAnyCPU               = flag.Bool("anycpu", false, "Do not use \"architecture\" of the manifest")
	flagExtractDir           = flag.Bool("p", false, "Specify the parent directory of *.exe into \"extract_dir\" and the basename into \"bin\"")
	flag32                   = flag.String("32", "386,486,586,686,32bit,win32", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 32bit")
	flag64                   = flag.String("64", "amd64,64bit,win64,x86_64", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 64bit")
	flagLicense              = flag.String("license", "", "Set the value of \"license\" of the manifest")
	flagDescription          = flag.String("description", "", "Set the value of \"description\" of the manifest")
	flagDownloadTo           = flag.String("downloadto", "", "Do not remove the downloaded zip files and save them onto the specified directory")
	flagBinPattern           = flag.String("binpattern", "*.exe", "The pattern for executables(separated with comma)")
)

func seekAssets(releases []*github.Release, fname string) (string, string) {
	for _, rel := range releases {
		for _, a := range rel.Assets {
			if a.Name == fname {
				return a.BrowserDownloadUrl, rel.TagName
			}
		}
	}
	return "", ""
}

func getHash(fname string) (string, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	h := sha256.New()
	io.Copy(h, fd)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type Archtecture struct {
	Url        string `json:"url"`
	Hash       string `json:"hash,omitempty"`
	ExtractDir string `json:"extract_dir,omitempty"`
}

type AutoUpdate struct {
	Archtectures map[string]*Archtecture `json:"architecture,omitempty"`
	UrlForAnyCPU string                  `json:"url,omitempty"`
}

type Manifest struct {
	Version       string                  `json:"version"`
	Description   string                  `json:"description,omitempty"`
	Homepage      string                  `json:"homepage,omitempty"`
	License       string                  `json:"license,omitempty"`
	Archtectures  map[string]*Archtecture `json:"architecture,omitempty"`
	UrlForAnyCPU  string                  `json:"url,omitempty"`
	HashForAnyCPU string                  `json:"hash,omitempty"`
	Bin           []string                `json:"bin"`
	CheckVer      string                  `json:"checkver"`
	AutoUpdate    AutoUpdate              `json:"autoupdate"`
}

func getBits(s string) string {
	s = strings.ToLower(s)
	for _, keyword := range strings.Split(*flag32, ",") {
		if strings.Contains(s, keyword) {
			return "32bit"
		}
	}
	for _, keyword := range strings.Split(*flag64, ",") {
		if strings.Contains(s, keyword) {
			return "64bit"
		}
	}
	if strings.Contains(s, "arm64") {
		return "arm64"
	}
	return ""
}

func listUpExeInZip(fname string, exeFiles map[string]struct{}) (string, error) {
	zr, err := zip.OpenReader(fname)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	patterns := strings.Split(strings.ToLower(*flagBinPattern), ",")

	var extractDir string
	for _, f := range zr.File {
		lowerName := strings.ToLower(f.Name)
		for _, pattern := range patterns {
			if matched, err := filepath.Match(pattern, lowerName); err == nil && matched {
				nm := f.Name
				if *flagExtractDir {
					notDir := filepath.Dir(nm)
					if notDir != "." {
						extractDir = notDir
						nm = filepath.Base(nm)
					}
				}
				exeFiles[nm] = struct{}{}
				break
			}
		}
	}
	return extractDir, nil
}

func getNameAndRepo() (string, string, error) {
	if *flagUserAndRepo != "" {
		var found bool
		name, repo, found := strings.Cut(*flagUserAndRepo, "/")
		if !found {
			return "", "", fmt.Errorf("-g \"%s\": format must be \"USER/REPO\"",
				*flagUserAndRepo)
		}
		return name, repo, nil
	}
	return gitdir.GetNameAndRepo()
}

func writeWithCRLF(source []byte, w io.Writer) error {
	for {
		before, after, found := bytes.Cut(source, []byte{'\n'})
		_, err := w.Write(before)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte{'\r', '\n'})
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		source = after
	}
}

var ignoreWords = flag.String("ignore", "linux,macos,freebsd,netbsd,darwin,plan9", "ignore the zipfile whose name contains these words")

func isWindowsZipName(name string) bool {
	for _, word := range strings.Split(*ignoreWords, ",") {
		if strings.Contains(name, word) {
			return false
		}
	}
	return true
}

func mains(args []string) error {
	if len(args) <= 0 && !*flagDownloadLatestAssets && *flagDownloadTo == "" {
		flag.PrintDefaults()
		return nil
	}
	name, repo, err := getNameAndRepo()
	if err != nil {
		return err
	}
	if name == "" || repo == "" {
		return errors.New("getNameAndRepo: can not find remote repository")
	}
	//println("name:", name)
	//println("repo:", repo)

	releases, err := github.GetReleases(name, repo, os.Stderr)
	if err != nil {
		return fmt.Errorf("getReleases: %w", err)
	}
	arch := make(map[string]*Archtecture)
	var url, tag string

	var binfiles = map[string]struct{}{}
	if len(args) > 0 {
		for _, arg1 := range args {
			files, err := filepath.Glob(arg1)
			if err != nil {
				files = []string{arg1}
			}
			for _, fname := range files {
				var bits string
				if !*flagAnyCPU {
					bits = getBits(fname)
					if bits == "" {
						return fmt.Errorf("%s: can not find `386` nor `amd64` nor `arm64`", fname)
					}
				}
				if !strings.EqualFold(filepath.Ext(fname), ".zip") {
					return fmt.Errorf("%s: not zipfile", fname)
				}
				name := filepath.Base(fname)
				url, tag = seekAssets(releases, name)
				if url == "" {
					return fmt.Errorf("%s not found in remote repository", name)
				}
				fmt.Fprintln(os.Stderr, "Read local file:", fname)
				hash, err := getHash(fname)
				if err != nil {
					return err
				}
				extractDir, err := listUpExeInZip(fname, binfiles)
				if err != nil {
					return err
				}
				arch[bits] = &Archtecture{
					Url:        url,
					Hash:       hash,
					ExtractDir: extractDir,
				}
			}
		}
	} else if *flagDownloadLatestAssets || *flagDownloadTo != "" {
		if len(releases) < 1 {
			return fmt.Errorf("%s/%s: no releases", name, repo)
		}
		for _, asset1 := range releases[0].Assets {
			name := asset1.Name
			if !strings.EqualFold(filepath.Ext(name), ".zip") {
				continue
			}
			if !isWindowsZipName(name) {
				continue
			}
			var bits string
			if !*flagAnyCPU {
				bits = getBits(name)
				if bits == "" {
					continue
				}
			}
			url = asset1.BrowserDownloadUrl
			fmt.Fprintln(os.Stderr, "Download:", url)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("%s: %w", url, err)
			}
			var tmpFd *os.File
			var tmpZipName string
			if *flagDownloadTo != "" {
				tmpZipName = filepath.Join(*flagDownloadTo, name)
				tmpFd, err = os.Create(tmpZipName)
				if err != nil {
					resp.Body.Close()
					return err
				}
				fmt.Fprintln(os.Stderr, "Save as", tmpZipName)
			} else {
				tmpFd, err = os.CreateTemp("", "make-scoop-manifest-*.zip")
				if err != nil {
					resp.Body.Close()
					return err
				}
				tmpZipName = tmpFd.Name()
				defer os.Remove(tmpZipName)
			}
			tag = releases[0].TagName

			io.Copy(tmpFd, resp.Body)
			resp.Body.Close()

			tmpFd.Seek(0, 0)
			h := sha256.New()
			io.Copy(h, tmpFd)
			hash := fmt.Sprintf("%x", h.Sum(nil))
			tmpFd.Close()

			extractDir, err := listUpExeInZip(tmpZipName, binfiles)
			if err != nil {
				return err
			}
			arch[bits] = &Archtecture{
				Url:        asset1.BrowserDownloadUrl,
				Hash:       hash,
				ExtractDir: extractDir,
			}
		}
	}

	var input []byte

	if *flagInlineTemplate != "" {
		input = []byte(*flagInlineTemplate)
	} else if *flagStdinTemplate {
		input, err = io.ReadAll(os.Stdin)
		if err != nil && err != io.EOF {
			return err
		}
	}
	var manifest Manifest
	if input != nil {
		if err = json.Unmarshal(input, &manifest); err != nil {
			return err
		}
	}
	if *flagDescription != "" {
		manifest.Description = *flagDescription
	}
	if *flagLicense != "" {
		manifest.License = *flagLicense
	}
	if binfiles != nil {
		for exe := range binfiles {
			manifest.Bin = append(manifest.Bin, exe)
		}
		sort.Strings(manifest.Bin)
	}
	if manifest.Archtectures == nil && !*flagAnyCPU {
		manifest.Archtectures = make(map[string]*Archtecture)
	}
	if manifest.AutoUpdate.Archtectures == nil {
		manifest.AutoUpdate.Archtectures = map[string]*Archtecture{}
	}
	if manifest.Homepage == "" {
		manifest.Homepage = fmt.Sprintf(
			"https://github.com/%s/%s", name, repo)
	}
	if manifest.CheckVer == "" {
		manifest.CheckVer = "github"
	}
	if *flagAnyCPU {
		manifest.Version = strings.TrimPrefix(tag, "v")
		manifest.UrlForAnyCPU = arch[""].Url
		manifest.HashForAnyCPU = arch[""].Hash
		manifest.AutoUpdate.UrlForAnyCPU =
			strings.ReplaceAll(arch[""].Url, manifest.Version, "$version")
	} else {
		for name, val := range arch {
			manifest.Archtectures[name] = val
			manifest.Version = strings.TrimPrefix(tag, "v")

			autoupdate := strings.ReplaceAll(val.Url, manifest.Version, "$version")
			bits := getBits(val.Url)
			manifest.AutoUpdate.Archtectures[bits] = &Archtecture{Url: autoupdate}
		}
	}
	if desc, err := github.GetDescription(name, repo, os.Stderr); err == nil {
		if manifest.Description == "" {
			description := desc.Description
			if description == "" {
				fmt.Fprintln(os.Stderr, `Warning: "description" field on GitHub is empty`)
			}
			manifest.Description = description
		}
		if manifest.License == "" {
			license := desc.License["name"]
			if license == "" {
				fmt.Fprintln(os.Stderr, `Warning: "license" field on GitHub is empty`)
			}
			manifest.License = license
		}
	}

	jsonBin, err := json.MarshalIndent(&manifest, "", "    ")
	if err != nil {
		return err
	}
	return writeWithCRLF(jsonBin, os.Stdout)
}

var version string

func main() {
	flag.Parse()

	fmt.Fprintf(os.Stderr, "%s %s for %s/%s by %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())

	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
