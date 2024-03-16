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
