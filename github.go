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
	"sort"
	"strings"

	"github.com/hymkor/make-scoop-manifest/internal/gitdir"
	"github.com/hymkor/make-scoop-manifest/internal/github"
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
	CheckVer      string                  `json:"checkver,omitempty"`
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

func isWindowsZipName(name string) bool {
	for _, word := range strings.Split(*flagIgnoreWords, ",") {
		if strings.Contains(name, word) {
			return false
		}
	}
	return true
}

type downloadAsset struct {
	zipName string
	hash    string
	tag     string
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

func tryGithub(args []string) error {
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

			arch[bits], err = downloadAndGetArchitecture(url, name, binfiles)
			if err != nil {
				return err
			}
			tag = releases[0].TagName
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
	if !*flagNoAutoUpdate {
		manifest.AutoUpdate = &AutoUpdate{}
	}
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
	if manifest.AutoUpdate != nil && manifest.AutoUpdate.Archtectures == nil {
		manifest.AutoUpdate.Archtectures = map[string]*Archtecture{}
	}
	if manifest.Homepage == "" {
		manifest.Homepage = fmt.Sprintf(
			"https://github.com/%s/%s", name, repo)
	}
	if !*flagNoAutoUpdate && manifest.CheckVer == "" {
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
