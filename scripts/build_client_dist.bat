@echo off

set GOOS=windows
set GOARCH=386

go build -ldflags "-s -w" -o .\build\client\client.exe .\client\

cd client-ui-qt
set SHITAMA_BUILD_ID=%1
set SHITAMA_COMMIT=%2
qmake
mingw32-make
mingw32-make clean
cd ..

mkdir .\dist\Shitama\
xcopy /y .\build\client\client.exe .\dist\Shitama\
xcopy /y .\build\client-ui-qt\Shitama.exe .\dist\Shitama\
windeployqt --no-system-d3d-compiler --no-angle --no-opengl-sw --no-svg --no-translations .\dist\Shitama\Shitama.exe
7z a .\dist\Shitama.7z .\dist\Shitama\
