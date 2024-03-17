make-scoop-manifest.exe
=======================

This is is the tool to make the manifest file of the [scoop-installer](https://scoop.sh) for your applications.

- Read releases information with GitHub-API
- The zip files' names must contain the word: `32bit`, `64bit`, `386`, `486`, `586`, `686`, `amd64`, `x86_64` or `arm64`
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

### hymkor/Download-Count.ps1 (PowerShell package)

```
make-scoop-manifest.exe -D -g hymkor/Download-Count.ps1 -binpattern "*.ps1" -anycpu > Download-Count.ps1.json
```

### mattn/bsky

```
make-scoop-manifest.exe -license MIT -D -g mattn/bsky -64 "" > bsky.json
```

There are only 64bit packages in the releases page, therefore we should give `-64 ""` as an option to regard `bsky-windows-X.Y.Z.zip` as 64bit.

### vim nightly build

```
make-scoop-manifest.exe -32 "x86" -64 "x64" -license Vim -D -ignore "pdb" -g vim/vim-win32-installer > vim-nightly.json
```

It installs from https://github.com/vim/vim-win32-installer/releases

### vim latest stable (experimentally)

- Get the URLs of ZIP files from the HTML of https://www.vim.org/download.php
- Only `vim.exe` and `gvim.exe` are installed on `scoop install`
- Registries are not modified on `scoop install`
- Some JSON-fields are gived with template from PowerShell Script ([examples/up-vim.ps1])

[examples/up-vim.ps1]: ./examples/up-vim.ps1

```examples\up-vim.ps1
$template = @'
{
    "checkver":{
        "regex":"gvim_(?<version>(?<major>\\d+)\\.(?<minor>\\d+)\\.\\d{1,3})_x64_signed\\.zip"
    },
    "autoupdate":{
        "architecture":{
            "64bit":{
                "url":"https://github.com/vim/vim-win32-installer/releases/download/v$version/gvim_$version_x64_signed.zip",
                "extract_dir":"vim\\vim$major$minor"
            },
            "32bit":{
                "url":"https://github.com/vim/vim-win32-installer/releases/download/v$version/gvim_$version_x86_signed.zip",
                "extract_dir":"vim\\vim$major$minor"
            }
        }
    }
}
'@
.\make-scoop-manifest.exe -p -inline $template -license Vim -binpattern "*vim.exe" -fromhtml https://www.vim.org/download.php > vim-hogehoge.json
```
