@setlocal
@set "PROMPT=$$ "
@call :test
@set "ERR=%ERRORLEVEL%"
@if exist tmp.json @del tmp.json
@endlocal
@exit /b %ERR%

:test
@echo ***
@echo *** DOWNLOAD TEST
@echo ***
make-scoop-manifest.exe hymkor/make-scoop-manifest > tmp.json
fc tmp.json make-scoop-manifest.json
@
@echo ***
@echo *** LOCAL ZIP TEST
@echo ***
make-scoop-manifest.exe .\dist\make-scoop-manifest-*-windows-*.zip > tmp.json
fc tmp.json make-scoop-manifest.json
