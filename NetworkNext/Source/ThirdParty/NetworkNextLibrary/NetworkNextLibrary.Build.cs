/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following 
    conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions 
       and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote 
       products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. 
    IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; 
    OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING 
    NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

using System.IO;
using UnrealBuildTool;

public class NetworkNextLibrary : ModuleRules
{
	public NetworkNextLibrary(ReadOnlyTargetRules Target) : base(Target)
	{
		Type = ModuleType.External;

        PublicIncludePaths.Add(Path.Combine(ModuleDirectory, "next", "include"));

        if (Target.Platform == UnrealTargetPlatform.Win64)
        {
            PublicLibraryPaths.Add(Path.Combine(ModuleDirectory, "next", "lib", "Win64", "Release"));
            PublicAdditionalLibraries.Add("next-win64-3.4.5.lib");
            PublicDelayLoadDLLs.Add("next-win64-3.4.5.dll");

            RuntimeDependencies.Add(Path.Combine(ModuleDirectory, "next", "lib", "Win64", "Release", "next-win64-3.4.5.dll"), StagedFileType.NonUFS);

            // This makes the editor work, because RuntimeDependencies do not apply to builds of the editor.
            Directory.CreateDirectory(Path.Combine(ModuleDirectory, "..", "..", "..", "Binaries", "Win64"));
            try
            {
                File.Copy(
                    Path.Combine(ModuleDirectory, "next", "lib", "Win64", "Release", "next-win64-3.4.5.dll"),
                    Path.Combine(ModuleDirectory, "..", "..", "..", "Binaries", "Win64", "next-win64-3.4.5.dll"),
                    true
                );
            }
            catch (System.IO.IOException)
            {
            }
        }
        else if (Target.Platform == UnrealTargetPlatform.Win32)
        {
            PublicLibraryPaths.Add(Path.Combine(ModuleDirectory, "next", "lib", "Win32", "Release"));
            PublicAdditionalLibraries.Add("next-win32-3.4.5.lib");
            PublicDelayLoadDLLs.Add("next-win32-3.4.5.dll");

            RuntimeDependencies.Add(Path.Combine(ModuleDirectory, "next", "lib", "Win32", "Release", "next-win32-3.4.5.dll"), StagedFileType.NonUFS);

            // This makes the editor work, because RuntimeDependencies do not apply to builds of the editor.
            Directory.CreateDirectory(Path.Combine(ModuleDirectory, "..", "..", "..", "Binaries", "Win32"));
            try
            {
                File.Copy(
                    Path.Combine(ModuleDirectory, "next", "lib", "Win32", "Release", "next-win32-3.4.5.dll"),
                    Path.Combine(ModuleDirectory, "..", "..", "..", "Binaries", "Win32", "next-win32-3.4.5.dll"),
                    true
                );
            }
            catch (System.IO.IOException)
            {
            }
        }
        else if (Target.Platform == UnrealTargetPlatform.XboxOne)
        {
            PublicAdditionalLibraries.Add(Path.Combine(ModuleDirectory, "next", "lib", "XboxOne", "Release", "next-xboxone-3.4.5.lib"));
            RuntimeDependencies.Add(Path.Combine(ModuleDirectory, "next", "lib", "XboxOne", "Release", "next-xboxone-3.4.5.dll"), StagedFileType.NonUFS);
        }
        else if (Target.Platform == UnrealTargetPlatform.PS4)
        {
            PublicAdditionalLibraries.Add(Path.Combine(ModuleDirectory, "next", "lib", "Playstation4", "Release", "next-ps4-3.4.5_stub.a"));
            RuntimeDependencies.Add(new RuntimeDependency(Path.Combine(ModuleDirectory, "next", "lib", "Playstation4", "Release", "next-ps4-3.4.5.prx")));
        }
        else if (Target.Platform == UnrealTargetPlatform.Switch)
        {
            PublicAdditionalLibraries.Add(Path.Combine(ModuleDirectory, "next", "lib", "NintendoSwitch-NX64", "Release", "next-nx64-3.4.5.nro"));
        }
    }
}
