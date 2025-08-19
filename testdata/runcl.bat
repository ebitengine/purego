echo off

REM @todo(judah): PATH SHOULD NOT BE HARDCODED, BETTER WAY TO FETCH THIS DYNAMICALLY?

call "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Auxiliary\\Build\\vcvarsall.bat" %1

cl.exe /LD /W1 /WX /Fe:%2 %3 %4 %5 %6 %7 %8 %9

del *.obj
