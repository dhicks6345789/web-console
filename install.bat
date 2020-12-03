@echo off
rem if "%1"=="" (
rem   echo How to use...
rem   goto end
rem )

echo Starting install...

set VERSION=0.1-alpha
set key=
set subdomain=

rem Parse any parameters.
:paramLoop
if "%1"=="" goto paramContinue
if "%1"=="--key" (
  shift
  set key=%2
)
if "%1"=="--subdomain" (
  shift
  set subdomain=%2
)
shift
goto paramLoop
:paramContinue

rem Stop any existing running services.
net stop WebConsole > nul 2>&1
net stop TunnelTo > nul 2>&1

echo Downloading Web Console v%VERSION% binary...
powershell -command "& {&'Invoke-WebRequest' -Uri https://github.com/dhicks6345789/web-console/releases/download/v%VERSION%/win-amd64.exe -OutFile 'C:\Program Files\WebConsole\webconsole.exe'}"
echo Downloading Web Console v%VERSION% support files...
powershell -command "& {&'Invoke-WebRequest' -Uri https://github.com/dhicks6345789/web-console/archive/v%VERSION%.zip -OutFile supportFiles.zip}"
rem xcopy /E /Y www "C:\Program Files\WebConsole\www" > nul 2>&1

rem copy tunnelto\tunnelto.exe "C:\Program Files\WebConsole" > nul 2>&1
rem mkdir "C:\Program Files\WebConsole\www" > nul 2>&1
rem mkdir "C:\Program Files\WebConsole\tasks" > nul 2>&1

rem Set up the WebConsole service.
rem nssm-2.24\win64\nssm install WebConsole "C:\Program Files\WebConsole\webconsole.exe" > nul 2>&1
rem nssm-2.24\win64\nssm set WebConsole DisplayName "Web Console" > nul 2>&1
rem nssm-2.24\win64\nssm set WebConsole AppNoConsole 1 > nul 2>&1
rem nssm-2.24\win64\nssm set WebConsole Start SERVICE_AUTO_START > nul 2>&1
net start WebConsole

exit /b 0

rem Allow the WebConsole service through the (local) Windows firewall.
netsh.exe advfirewall firewall add rule name="WebConsole" program="C:\Program Files\WebConsole\webconsole.exe" protocol=tcp dir=in enable=yes action=allow profile="private,domain,public" > nul 2>&1

if not "%key%"=="" (
  if not "%subdomain%"=="" (
    rem Set up the TunnelTo.dev service.
    nssm-2.24\win64\nssm install TunnelTo "C:\Program Files\WebConsole\tunnelto.exe" --port 8090 --key %key% --subdomain %subdomain% > nul 2>&1
    nssm-2.24\win64\nssm set TunnelTo DisplayName "TunnelTo.dev" > nul 2>&1
    nssm-2.24\win64\nssm set TunnelTo AppNoConsole 1 > nul 2>&1
    nssm-2.24\win64\nssm set TunnelTo Start SERVICE_AUTO_START > nul 2>&1
    net start TunnelTo
  )
)

:end
