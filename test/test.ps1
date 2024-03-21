$testset = @(
    @('-g benhoyt/goawk -D',
        'goawk.json' ),
    @('-anycpu -g zat-kaoru-hayama/yShowVer -D',
        'yShowVer.json'),
    @('-p -license "MIT License" -g mattn/twty -D',
        'twty.json'),
    @('-D -g hymkor/Download-Count.ps1 -binpattern "*.ps1" -anycpu',
        "Download-Count.ps1.json"),
    @('-license MIT -D -g mattn/bsky -64 ""',
        'bsky.json')
)

$success = 0
$failure = 0

foreach($p in $testset){
    $expect = $p[1]
    $result = (Join-Path $env:TEMP $p[1])

    Write-Host "expect:" $expect
    Write-Host "result:" $result

    $commandline = ("..\make-scoop-manifest " + $p[0])
    Invoke-expression $commandline > $result
    if ( $LastExitCode -ne 0 ){
        break
    }
    fc.exe $expect $result
    if ( $LastExitCode -ne 0 ){
        Write-Error ("Test error for " + $p[1])
        $failure++
    } else {
        $success++
    }
}

Write-Host ("{0} of {1} tests succeeded" -f ($success, $success+$failure ))
Write-Host ("{0} of {1} tests failed"    -f ($failure, $success+$failure ))
