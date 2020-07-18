@echo off

nssm-2.24\win64\nssm install WebConsole "C:\Program Files\WebConsole\webconsole.exe"
nssm-2.24\win64\nssm set WebConsole DisplayName "Web Console"
nssm-2.24\win64\nssm set WebConsole AppNoConsole 1
nssm-2.24\win64\nssm set WebConsole Start SERVICE_AUTO_START
