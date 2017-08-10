@echo off
del *.exe /q  2>nul
del *.log /q  2>nul
del temp\*.* /q 2>nul
rmdir temp 2>nul
git add *
git commit * -m''
git push

