To compile linux binaries from a windows machine:

1. Run generateSymLink.bat through command prompt with admin privledges
2. Install Ubuntu 18.04 from windows store
3. Configure visual studio to connect to it https://devblogs.microsoft.com/cppblog/targeting-windows-subsystem-for-linux-from-visual-studio/
4. Open Visual Studio, the click File -> Open -> CMake -> CMakeLists.txt
5. Edit outputDirRoot in CMakeLists.txt to point to desired binary output location
6. Select your desired config in Visual Studio (Linux-Debug or Linux-Release), then click CMake -> Rebuild All