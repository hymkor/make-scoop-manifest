make-scoop-manifest.exe
=======================

This tool generates a Scoop manifest file for your application using information from GitHub Releases.

- Your application must be packaged as a zip file and attached as an asset on GitHub Releases.
- The names of the zip files must contain one of the following keywords:
  `32bit`, `64bit`, `386`, `486`, `586`, `686`, `amd64`, `x86_64`, `x64`, or `arm64`

  - If the executable is AnyCPU, use the `-anycpu` option.
- If the filenames contain `linux` or `macos`, those assets are ignored.
- The tool does not check whether an existing manifest is up to date.

A sample output is available [here](https://github.com/hymkor/make-scoop-manifest/blob/master/make-scoop-manifest.json).

Install
-------

```
scoop bucket add hymkor https://github.com/hymkor/scoop-bucket
scoop install make-scoop-manifest
```

OR

```
scoop install https://raw.githubusercontent.com/hymkor/make-scoop-manifest/master/make-scoop-manifest.json
```

Usage
-----

```
make-scoop-manifest {-options} [REPOSITORY] {localfiles...} > MANIFEST.JSON
```

+ **REPOSITORY** — `"OWNER/REPO"` or a GitHub URL
  Examples:

  + `hymkor/make-scoop-manifest`
  + `https://github.com/hymkor/make-scoop-manifest`
  + `git@github.com:hymkor/make-scoop-manifest.git`
    If omitted, the repository information is determined via `git remote show`.
+ **localfiles**
  If specified, the given local zip files are used as assets instead of downloading from GitHub.

> [!Note]
> As of v0.10.0, the `-g` and `-D` options are optional.

Example 1
---------

Retrieve all information from a GitHub repository:

```
$ make-scoop-manifest.exe hymkor/make-scoop-manifest  1>tmp.json

make-scoop-manifest.exe v0.9.0-29-ge4752cc for windows/amd64 by go1.22.1
Owner: hymkor
Repos: make-scoop-manifest
Get: https://api.github.com/repos/hymkor/make-scoop-manifest/releases
Download: https://github.com/hymkor/make-scoop-manifest/releases/download/v0.9.0/make-scoop-manifest-v0.9.0-windows-386.zip
Download: https://github.com/hymkor/make-scoop-manifest/releases/download/v0.9.0/make-scoop-manifest-v0.9.0-windows-amd64.zip
Get: https://api.github.com/repos/hymkor/make-scoop-manifest
```

Example 2
---------

When the repository is not specified, information is detected via `git remote show`.
Local zip files are used to generate `"hash"` and `"bin"`.

```
$ cd %USERPROFILE%\src\make-scoop-manifest
$ make-scoop-manifest.exe .\dist\make-scoop-manifest-*-windows-*.zip  1>tmp.json

make-scoop-manifest.exe v0.9.0-30-gad96d30 for windows/amd64 by go1.22.1
[git remote show]
> origin
[git remote show -n origin]
> * remote origin
>   Fetch URL: git@github.com:hymkor/make-scoop-manifest.git
>   Push  URL: git@github.com:hymkor/make-scoop-manifest.git
Owner: hymkor
Repos: make-scoop-manifest
Get: https://api.github.com/repos/hymkor/make-scoop-manifest/releases
Read local file: dist\make-scoop-manifest-v0.9.0-windows-386.zip
Read local file: dist\make-scoop-manifest-v0.9.0-windows-amd64.zip
Get: https://api.github.com/repos/hymkor/make-scoop-manifest
```

Sample command-line options
---------------------------

### benhoyt/goawk

```
make-scoop-manifest benhoyt/goawk > bucket/goawk.json
```

### zat-kaoru-hayama/yShowver

```
make-scoop-manifest.exe -anycpu zat-kaoru-hayama/yShowVer > bucket/yShowVer.json 
```

### mattn/twty

```
make-scoop-manifest -p -license "MIT License" mattn/twty > bucket/twty.json
```

### hymkor/Download-Count.ps1 (PowerShell package)

```
make-scoop-manifest.exe -binpattern "*.ps1" -anycpu hymkor/Download-Count.ps1 > Download-Count.ps1.json
```

### mattn/bsky

```
make-scoop-manifest.exe -license MIT -64 "" mattn/bsky > bsky.json
```

Only 64-bit packages are available for this repository, so `-64 ""` is required to treat `bsky-windows-X.Y.Z.zip` as a 64-bit package.
