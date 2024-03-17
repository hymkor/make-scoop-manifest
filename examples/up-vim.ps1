$template = @'
{
    "bin":[
        "vim.exe",
        [ "vim.exe", "vi" ],
        [ "vim.exe", "ex", "-e" ],
        [ "vim.exe", "view", "-R" ],
        [ "vim.exe", "rvim", "-Z" ],
        [ "vim.exe", "rview", "-RZ" ],
        [ "vim.exe", "vimdiff", "-d" ],
        "gvim.exe",
        [ "gvim.exe", "gview", "-R" ],
        [ "gvim.exe", "evim", "-y" ],
        [ "gvim.exe", "eview", "-Ry" ],
        [ "gvim.exe", "rgvim", "-Z" ],
        [ "gvim.exe", "rgview", "-RZ" ],
        [ "gvim.exe", "gvimdiff", "-d" ],
        "xxd.exe"
    ],
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
make-scoop-manifest.exe -p -inline $template -license Vim -fromhtml https://www.vim.org/download.php > vim-hogehoge.json
