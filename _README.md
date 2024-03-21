make-scoop-manifest.exe
=======================

This is a tool to make the manifest file of the [scoop-installer](https://scoop.sh) for your application on GitHub Releases.

- Your application must be packaged as a zip-file and attached as an asset in GitHub Releases
- The names of zip-files names must contain the word: `32bit`, `64bit`, `386`, `486`, `586`, `686`, `amd64`, `x86_64`, `x64` or `arm64`
    - If the executable is for AnyCPU, use the option `-anycpu`.
- If the names of zip-files contain `linux` or `macos`, they are ignored.
- Do not check the target is updated or not.

The output sample is [here](https://github.com/hymkor/make-scoop-manifest/blob/master/make-scoop-manifest.json).

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

+ REPOSITORY - "OWNERNAME/REPOSITORY" or GitHub-URL
    + For example:
        + `hymkor/make-scoop-manifest`
        + `https://github.com/hymkor/make-scoop-manifest`
    + If omitted, get them with `git remote show`
+ localfiles
    + If given, use the localfiles as assets instead of downloading

> [!Note]
> The option `-g` and `-D` can be omitted now since v0.10.0

Example-1
---------

- Get all information from GitHub Repository

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

Example-2
---------

- When REPOSITORY is not specified, get information about repository with `git remote show`.
- Make "hash" and "bin" of the manifest file with reading the local-zip files.

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

Sample commandline options:
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

There are only 64bit packages in the releases page, therefore we should give `-64 ""` as an option to regard `bsky-windows-X.Y.Z.zip` as 64bit.
