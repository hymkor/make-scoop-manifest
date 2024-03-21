package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/hymkor/make-scoop-manifest/internal/flag"
	"github.com/hymkor/make-scoop-manifest/internal/gitdir"
	"github.com/hymkor/make-scoop-manifest/internal/github"
)

var (
	flagInlineTemplate = flag.String("inline", "", "Read the template of the manifest JSON from the argument")
	flagStdinTemplate  = flag.Bool("stdin", false, "Read the template of the manifest JSON from the standard input")
	flagAnyCPU         = flag.Bool("anycpu", false, "Do not use \"architecture\" of the manifest")
	flagExtractDir     = flag.Bool("p", false, "Specify the parent directory of *.exe into \"extract_dir\" and the basename into \"bin\"")
	flag32             = flag.String("32", "386,486,586,686,32bit,win32", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 32bit")
	flag64             = flag.String("64", "amd64,64bit,win64,x86_64,x64", "When anyone of the specified strings is found in the zipfile's name, judge its architecture is 64bit")
	flagLicense        = flag.String("license", "", "Set the value of \"license\" of the manifest")
	flagDescription    = flag.String("description", "", "Set the value of \"description\" of the manifest")
	flagDownloadTo     = flag.String("downloadto", "", "Do not remove the downloaded zip files and save them onto the specified directory")
	flagBinPattern     = flag.String("binpattern", "*.exe", "The pattern for executables(separated with comma)")
	flagIgnoreWords    = flag.String("ignore", "linux,macos,freebsd,netbsd,darwin,plan9", "ignore the zipfile whose name contains these words")
	flagNoAutoUpdate   = flag.Bool("noautoupdate", false, "disable autoupdate")
)

var (
	//Deprecated:
	_ = flag.Bool("D", false, "(deprecated) Download and read the latest assets from GitHub")
	//Deprecated:
	flagUserAndRepo = flag.String("g", "", "(deprecated) Specify GitHub's \"USER/REPOSITORY\"")
)

func readTemplate() (*Manifest, error) {
	var manifest Manifest
	var input []byte

	if *flagInlineTemplate != "" {
		input = []byte(*flagInlineTemplate)
	} else if *flagStdinTemplate {
		var err error
		input, err = io.ReadAll(os.Stdin)
		if err != nil && err != io.EOF {
			return nil, err
		}
	}
	if input != nil {
		if err := json.Unmarshal(input, &manifest); err != nil {
			return nil, err
		}
	}
	return &manifest, nil
}

// keysToSlice returns keys of map1 as the type []string or a string
func keysToSlice(map1 map[string]struct{}) any {
	if map1 == nil {
		return nil
	}
	if len(map1) == 1 {
		for binname := range map1 {
			return binname
		}
	}
	slice := make([]string, 0, len(map1))
	for key := range map1 {
		slice = append(slice, key)
	}
	sort.Strings(slice)
	return slice
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
	Bin           any                     `json:"bin"`
	CheckVer      any                     `json:"checkver,omitempty"`
	AutoUpdate    *AutoUpdate             `json:"autoupdate,omitempty"`
}

func getBits(s string) string {
	s = strings.ToLower(s)
	for _, keyword := range strings.Split(*flag64, ",") {
		if strings.Contains(s, keyword) {
			return "64bit"
		}
	}
	for _, keyword := range strings.Split(*flag32, ",") {
		if strings.Contains(s, keyword) {
			return "32bit"
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

func isWindowsZipName(name string) bool {
	for _, word := range strings.Split(*flagIgnoreWords, ",") {
		if strings.Contains(name, word) {
			fmt.Fprintf(os.Stderr, "%s: ignored because it contains %s\n", name, word)
			return false
		}
	}
	return true
}

type downloadAsset struct {
	zipName string
	hash    string
	url     string
}

func (d *downloadAsset) Dispose() {
	os.Remove(d.zipName)
}

func downloadAsTmpZip(url, name string) (*downloadAsset, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", url, err)
	}
	defer resp.Body.Close()

	var tmpFd *os.File
	var tmpZipName string
	if *flagDownloadTo != "" {
		tmpZipName = filepath.Join(*flagDownloadTo, name)
		tmpFd, err = os.Create(tmpZipName)
		if err != nil {
			return nil, err
		}
		fmt.Fprintln(os.Stderr, "Save as", tmpZipName)
	} else {
		tmpFd, err = os.CreateTemp("", "make-scoop-manifest-*.zip")
		if err != nil {
			return nil, err
		}
		tmpZipName = tmpFd.Name()
	}

	io.Copy(tmpFd, resp.Body)

	tmpFd.Seek(0, 0)
	h := sha256.New()
	io.Copy(h, tmpFd)
	hash := fmt.Sprintf("%x", h.Sum(nil))
	tmpFd.Close()

	return &downloadAsset{
		zipName: tmpZipName,
		hash:    hash,
		url:     url,
	}, nil
}

func readFileAndGetArchitecture(url, fullpath string, foundExecutables map[string]struct{}) (*Archtecture, error) {
	hash, err := getHash(fullpath)
	if err != nil {
		return nil, err
	}
	extractDir, err := listUpExeInZip(fullpath, foundExecutables)
	if err != nil {
		return nil, err
	}
	return &Archtecture{
		Url:        url,
		Hash:       hash,
		ExtractDir: extractDir,
	}, nil
}

func downloadAndGetArchitecture(url, name string, foundExecutables map[string]struct{}) (*Archtecture, error) {
	downloadZip, err := downloadAsTmpZip(url, name)
	if err != nil {
		return nil, err
	}
	defer downloadZip.Dispose()

	extractDir, err := listUpExeInZip(downloadZip.zipName, foundExecutables)
	if err != nil {
		return nil, err
	}

	return &Archtecture{
		Url:        url,
		Hash:       downloadZip.hash,
		ExtractDir: extractDir,
	}, nil
}

var (
	rxRepositoryG = regexp.MustCompile(`^git@github.com:([^/]+)/(.+)$`)
	rxRepositoryH = regexp.MustCompile(`^(?:https://github.com/)?([^/]+)/([^/]+)`)
)

func mains(args []string) error {
	// for compatiblity
	if *flagUserAndRepo != "" {
		args = slices.Insert(args, 0, *flagUserAndRepo)
	}
	localfiles := map[string]string{}
	var owner, repos string

	for _, arg1 := range args {
		if repos == "" && !strings.EqualFold(filepath.Ext(arg1), ".zip") {
			if m := rxRepositoryG.FindStringSubmatch(arg1); m != nil {
				owner = m[1]
				if strings.EqualFold(filepath.Ext(m[2]), ".git") {
					repos = m[2][:len(m[2])-4]
				} else {
					repos = m[2]
				}
				continue
			}
			if m := rxRepositoryH.FindStringSubmatch(arg1); m != nil {
				owner = m[1]
				repos = m[2]
				continue
			}
		}
		files, err := filepath.Glob(arg1)
		if err != nil {
			files = []string{arg1}
		}
		for _, fname := range files {
			name := filepath.Base(fname)
			localfiles[name] = fname
		}
	}
	if owner == "" {
		var err error
		owner, repos, err = gitdir.GetNameAndRepo(os.Stderr)
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "Owner:", owner)
	fmt.Fprintln(os.Stderr, "Repos:", repos)

	releases, err := github.GetReleases(owner, repos, os.Stderr)
	if err != nil {
		return fmt.Errorf("getReleases: %w", err)
	}
	if len(releases) < 1 {
		return fmt.Errorf("%s/%s: no releases", owner, repos)
	}
	fmt.Fprintln(os.Stderr, "Search the assets of", releases[0].TagName)

	arch := make(map[string]*Archtecture)
	var tag string

	var binfiles = map[string]struct{}{}

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
		url := asset1.BrowserDownloadUrl
		if fullpath, ok := localfiles[name]; ok {
			fmt.Fprintln(os.Stderr, "Read local file:", fullpath)
			arch[bits], err = readFileAndGetArchitecture(url, fullpath, binfiles)
		} else {
			fmt.Fprintln(os.Stderr, "Download:", url)
			arch[bits], err = downloadAndGetArchitecture(url, name, binfiles)
		}
		if err != nil {
			return err
		}
		tag = releases[0].TagName
	}

	manifest, err := readTemplate()
	if err != nil {
		return err
	}

	if !*flagNoAutoUpdate {
		manifest.AutoUpdate = &AutoUpdate{}
	}
	if *flagDescription != "" {
		manifest.Description = *flagDescription
	}
	if *flagLicense != "" {
		manifest.License = *flagLicense
	}
	if manifest.Bin == nil {
		manifest.Bin = keysToSlice(binfiles)
	}
	if manifest.Archtectures == nil && !*flagAnyCPU {
		manifest.Archtectures = make(map[string]*Archtecture)
	}
	if manifest.AutoUpdate != nil && manifest.AutoUpdate.Archtectures == nil {
		manifest.AutoUpdate.Archtectures = map[string]*Archtecture{}
	}
	if manifest.Homepage == "" {
		manifest.Homepage = fmt.Sprintf(
			"https://github.com/%s/%s", owner, repos)
	}
	if !*flagNoAutoUpdate && manifest.CheckVer == nil {
		manifest.CheckVer = "github"
	}
	if *flagAnyCPU {
		manifest.Version = strings.TrimPrefix(tag, "v")
		arch1 := arch[""]
		if arch1 == nil {
			return errors.New("assets not found")
		}
		manifest.UrlForAnyCPU = arch1.Url
		manifest.HashForAnyCPU = arch1.Hash
		if manifest.AutoUpdate != nil {
			manifest.AutoUpdate.UrlForAnyCPU =
				strings.ReplaceAll(arch1.Url, manifest.Version, "$version")
		}
	} else {
		for name, val := range arch {
			manifest.Archtectures[name] = val
			manifest.Version = strings.TrimPrefix(tag, "v")

			autoupdate := strings.ReplaceAll(val.Url, manifest.Version, "$version")
			bits := getBits(val.Url)
			if manifest.AutoUpdate != nil {
				manifest.AutoUpdate.Archtectures[bits] = &Archtecture{Url: autoupdate}
			}
		}
	}
	if desc, err := github.GetDescription(owner, repos, os.Stderr); err == nil {
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

	jsonBin, err := json.MarshalIndent(manifest, "", "    ")
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
