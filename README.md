make-scoop-manifest.exe
=======================

This is is the tool to make the manifest file of the [scoop-installer](https://scoop.sh) for your applications.

- Read releases information with GitHub-API
- The zip files' names must contain the word: `32bit`,`64bit`,`386`,`amd64` or `arm64`
    - If the executable is for AnyCPU, use the option `-anycpu`.
- If the zip files' names contains `linux` or `macos`, they are ignored.
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

Usage-1
-------

```
cd YOUR-REPOSITORY
make-scoop-manifest *.zip > YOUR-TOOL.json
```

- Get USERNAME and REPOSITORY with `git remote show`.
- Make "hash" and "bin" of the manifest file with reading the local-zip files.

Example:
```
$ ../make-scoop-manifest/make-scoop-manifest.exe ./*.zip > zar.json
Get: https://api.github.com/repos/hymkor/zar/releases
Read local file: zar-v0.2.2-windows-386.zip
Read local file: zar-v0.2.2-windows-amd64.zip
Get: https://api.github.com/repos/hymkor/zar
```

Usage-2
-------

```
make-scoop-manifest -g USERNAME/REPOSITORY *.zip > YOUR-TOOL.json
```

- Get USERNAME and REPOSITORY with the option parameter.
- Make "hash" and "bin" of the manifest file with reading the local-zip files.

```
$ make-scoop-manifest.exe -g hymkor/zar ../zar/*.zip > zar.json
Get: https://api.github.com/repos/hymkor/zar/releases
Read local file: ..\zar\zar-v0.2.2-windows-386.zip
Read local file: ..\zar\zar-v0.2.2-windows-amd64.zip
Get: https://api.github.com/repos/hymkor/zar
```

Usage-3
-------

```
make-scoop-manifest -D -g USERNAME/REPOSITORY > YOUR-TOOL.json
```

- Get USERNAME and REPOSITORY with the option parameter.
- Make "hash" and "bin" of the manifest file with downloading and reading the **uploaded zip files of the latest assets in the Releases**.  
  (Caution: the download counters are incremented)

Example:
```
$ make-scoop-manifest.exe -D -g hymkor/zar > zar.json
Get: https://api.github.com/repos/hymkor/zar/releases
Download: https://github.com/hymkor/zar/releases/download/v0.2.2/zar-v0.2.2-windows-386.zip
Download: https://github.com/hymkor/zar/releases/download/v0.2.2/zar-v0.2.2-windows-amd64.zip
Get: https://api.github.com/repos/hymkor/zar
```

Sample commandline options:
---------------------------

### benhoyt/goawk

```
make-scoop-manifest -g benhoyt/goawk -D > bucket/goawk.json
```

### zat-kaoru-hayama/yShowver

```
make-scoop-manifest.exe -anycpu -g zat-kaoru-hayama/yShowVer -D > bucket/yShowVer.json 
```

### mattn/twty

```
make-scoop-manifest -p -license "MIT License" -g mattn/twty -D > bucket/twty.json
```

### koron/vim-kaoriya

```
make-scoop-manifest -p -license "Vim License" -g koron/vim-kaoriya vim82-kaoriya-*.zip > vim-kaoriya.json
```
