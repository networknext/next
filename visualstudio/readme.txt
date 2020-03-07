This folder is intended to be shipped as part of plugin download in the following location:
\NetworkNext\Source\ThirdParty\NetworkNextLibrary\next\vs

If it is in the right place, and you have the appropriate SDKs for the platforms you would like to target, you should be able to open next.sln, then select Build -> Batch Build to rebuild the next sdk from source for the desired platforms, and have the binaries automatically placed in the folder UE4 will look for them (ThirdParty\NetworkNextLibrary\next\lib)

Building linux binaries on windows requires some extra steps, instructions on how to do this can be found under linux/readme.txt