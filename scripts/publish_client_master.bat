@echo off

set SHITAMA_QINIU_AK=%1
set SHITAMA_QINIU_SK=%2
set SHITAMA_BUILD_ID=%3

curl -o qshell.exe http://devtools.qiniu.com/2.0.8/qshell-windows-x86.exe
qshell account %SHITAMA_QINIU_AK% %SHITAMA_QINIU_SK%
qshell fput shitama Shitama-r%SHITAMA_BUILD_ID%.7z .\dist\Shitama.7z
qshell copy -overwrite shitama Shitama-r%SHITAMA_BUILD_ID%.7z shitama Shitama.7z
qshell cdnrefresh .\assets\cdnrefresh.qiniu.txt
