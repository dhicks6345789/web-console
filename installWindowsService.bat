@echo off

rem Set up the WebConsole service.
net stop WebConsole
nssm-2.24\win64\nssm install WebConsole "C:\Program Files\WebConsole\webconsole.exe"
nssm-2.24\win64\nssm set WebConsole DisplayName "Web Console"
nssm-2.24\win64\nssm set WebConsole AppNoConsole 1
nssm-2.24\win64\nssm set WebConsole Start SERVICE_AUTO_START
net start WebConsole

rem Allow the WbConsole service through the firewall.
netsh.exe advfirewall firewall add rule name="WebConsole" program="C:\Program Files\WebConsole\webconsole.exe" protocol=tcp dir=in enable=yes action=allow profile="private,domain,public"
