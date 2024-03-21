- フィールド AutoUpdate を出力しないオプション `-noautoupdate` を追加
- `x86_64` が32bit と誤判定される問題を修正するため、64bitパターンを32bitパターンより先に探すようにした。
- オプション`-64`のデフォルトに `x64` を追加
- JSONのテンプレート読み込みのため、`-inline` や `-stdin` が指定された時、フィールド`"checkver"`, `"bin"` には文字列だけでなく、任意のJSONオブジェクトを指定できるようにした
- `-D` は省略できるようにした。ローカルファイルとしてファイル名が与えられていない場合は、自動で GitHub からダウンロードするようにした
- `-g` は省略できるようにした。レポジトリ名はいきなりコマンドラインに書けるようにした。
- 実行可能ファイルが一つの時、`"bin"` フィールドは配列ではなく、1文字列で出力するようにした。
- `-ignore` のキーワードが理由で無視された Assets はログに出力するようにした

v0.9.0
======
2024-01-27

- `git remote show` でレポジトリの場所を得る場合、ssh プロトコルでクローンしたレポジトリしかサポートしていなかったが、https もサポートした。

v0.8.0
======
2024-01-27

- 実行ファイルのパターンを指定するオプション `-binpattern PATTERN` を追加 （例：`-binpattern "*.exe,*.ps1,*.cmd"`）
- レポジトリ名が `.` (ドット) を含んでいる時にレポジトリの検出ができない問題を修正
- 与えられたパラメータの添付ファイル名が見つからなかった時、パニックが起きる問題を修正

PowerShell のレポジトリのマニフェストを作るためには、次のように呼び出します。

```
	make-scoop-manifest -binpattern "*.ps1" -anycpu $(NAME)-*.zip > $(NAME).json
```

v0.7.0
======
2023-10-26

- Rustアプリケーション向けにCPU アーキテクチャを判断するキーワードを追加
    - 32bit： `486`, `586`, および `686`
    - 64bit： `x86_64`

v0.6.0
======
(2023-01-25)

- オプションや ZIP ファイルが与えられなかったとき、ヘルプを表示するようにした
- オプション  `-downloadto DIRECTORY` を追加（`-D`と等価だが、ZIPファイルを削除せず、DIRECTORYに残す)
- "bin" の項目をソートするようにした
- オプション `-license` を追加 (  `-license "MIT"` は `-inline "{ \"license\":\"MIT\" }"` と等価となる )
- オプション `-description` を追加 (  `-description "XXX"` は `-inline "{ \"description\":\"XXX\" }"` と等価となる )

v0.5.0
=======
2023-01-17

- `-32 "string"` ：指定された文字列が見つかった時、architecture を32bitにする（デフォルト `"386,32bit,win32"`）
- `-64 "string"` ：指定された文字列が見つかった時、architecture を64bitにする（デフォルト `"amd64,64bit,win64"`）

v0.4.0
=======
2023-01-17

- オプション `-P` を追加
    - \*.exe の親ディレクトリを `"extract_dir"` に設定する
    - \*.exe のファイル名部分を `"bin"` に設定する

v0.3.0
=======
2023-01-12

- `-anycpu` オプションを追加
    - `"architecture"` フィールドを使わない
    - ZIPファイル名がキーワード：`386` や `amd64` を持っているかのチェックをしない

オプション `-anycpu` が使われたときの出力サンプルは次のとおり

```
{
    "version": "2.0.0.6",
    "description": "Show the version number , timestamp and MD5SUM of Windows Executables",
    "homepage": "https://github.com/zat-kaoru-hayama/yShowVer",
    "license": "MIT License",
    "url": "https://github.com/zat-kaoru-hayama/yShowVer/releases/download/v2.0.0.6/yShowVer-2.0.0.6.zip",
    "hash": "1b6f01937d771fae71546e6d94296dcdedcd00c28723b7185ed54082f193f374",
    "bin": [
        "yShowVer.exe"
    ],
    "checkver": "github",
    "autoupdate": {
        "url": "https://github.com/zat-kaoru-hayama/yShowVer/releases/download/v$version/yShowVer-$version.zip"
    }
}
```

v0.2.2
=======
2023-01-10

- `"license"` や `"description"` が空の時、警告を表示するようにした

v0.2.1
=======
2023-01-09

- (#2) 項目 `"checkver"` には `"github"` だけを設定するようにした (  Thx @spiegel-im-spiegel )

v0.2.0
=======
2023-01-09

- (#1) アーキテクチャのキーワードとして `32bit`,`64bit`,`arm64` も利用可能になった (Thx @spiegel-im-spiegel )


v0.1.1
=======
2023-01-09

- レポジトリ名に `-` が含まれているときにエラーになる問題を修正

v0.1.0
=======
2023-01-09

- 初版
