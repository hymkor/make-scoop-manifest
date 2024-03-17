package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

var rxTitle = regexp.MustCompile(`(?is)<title>([^>]*)</title>`)

var rxAnchor = regexp.MustCompile(`(?is)<a[^>]+href="([^"]+)"`)

func tryFromHtml(downloadPageUrl string) error {
	rxVersion, err := regexp.Compile(*flagVersionPattern)
	if err != nil {
		return err
	}
	resp, err := http.Get(downloadPageUrl)
	if err != nil {
		return err
	}
	htmlBin, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	html := string(htmlBin)

	title := rxTitle.FindStringSubmatch(html)
	if len(title) >= 2 {
		fmt.Fprintln(os.Stderr, "Title:", title[1])
	}
	m := rxAnchor.FindAllStringSubmatch(html, -1)
	urls := make(map[string][]string)
	var version string
	for _, m1 := range m {
		linkUrl := m1[1]
		if strings.HasPrefix(linkUrl, "ftp:") {
			continue
		}
		if !strings.HasSuffix(linkUrl, ".zip") {
			continue
		}
		if !isWindowsZipName(linkUrl) {
			continue
		}
		if !strings.HasPrefix(linkUrl, "http") {
			if _linkUrl, err := url.JoinPath(downloadPageUrl, linkUrl); err == nil {
				linkUrl = _linkUrl
			}
		}
		if v := rxVersion.FindString(linkUrl); v == "" || v < version {
			continue
		} else {
			version = strings.TrimPrefix(v, "v")
		}
		if bitword := getBits(linkUrl); bitword != "" {
			urls[bitword] = append(urls[bitword], linkUrl)
			fmt.Fprintln(os.Stderr, "Found:", linkUrl, "as", bitword, "for", version)
		}
	}
	var errs []error
	for bitWord, linkUrls := range urls {
		if len(linkUrls) != 1 {
			errs = append(errs,
				fmt.Errorf("%s: multiple URLs found\n\t%s",
					bitWord,
					strings.Join(linkUrls, "\n\t")))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	manifest, err := readTemplate()
	if err != nil {
		return err
	}
	manifest.Version = version
	if manifest.Description == "" {
		manifest.Description = title[1]
	}
	if manifest.Homepage == "" {
		manifest.Homepage = downloadPageUrl
	}
	if *flagLicense != "" {
		manifest.License = *flagLicense
	}
	manifest.Archtectures = make(map[string]*Archtecture)
	binfiles := map[string]struct{}{}

	for bit, urls := range urls {
		for _, url1 := range urls {
			fmt.Fprintln(os.Stderr, "Download:", url1)
			name := path.Base(url1)
			arch, err := downloadAndGetArchitecture(url1, name, binfiles)
			if err != nil {
				return err
			}
			manifest.Archtectures[bit] = arch
		}
	}
	if manifest.Bin == nil {
		manifest.Bin = keysToSlice(binfiles)
	}
	jsonBin, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		return err
	}
	return writeWithCRLF(jsonBin, os.Stdout)
}
