/*
    Network Next SDK 3.4.0

    Copyright © 2017 - 2020 Network Next, Inc.

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

using UnrealBuildTool;

public class NetworkNext : ModuleRules
{
    public NetworkNext(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = ModuleRules.PCHUsageMode.UseExplicitOrSharedPCHs;

        Definitions.Add("NETWORKNEXT_PACKAGE=1");
        Definitions.Add("NETWORKNEXT_UNREAL_ENGINE_416=1");
        Definitions.Add("NEXT_SHARED=1");

        if (Target.Platform == UnrealTargetPlatform.Win64)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD_WIN64=1");
        }
        else if (Target.Platform == UnrealTargetPlatform.Win32)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD_WIN32=1");
        }
        else if (Target.Platform == UnrealTargetPlatform.Linux)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
        }
        // <XBOX
        else if (Target.Platform == UnrealTargetPlatform.XboxOne)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
        }
        // XBOX>
        // <PS4
        else if (Target.Platform == UnrealTargetPlatform.PS4)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD=1");
            Definitions.Add("NETWORKNEXT_ENABLE_DELAY_LOAD_PS4=1");
        }
        // PS4>
        // <SWITCH
        else if (Target.Platform == UnrealTargetPlatform.Switch)
        {
            Definitions.Add("NETWORKNEXT_AVAILABLE=1");
        }
        // SWITCH>

        PublicDependencyModuleNames.AddRange(
            new string[] {
                "OnlineSubsystemUtils",
                "NetworkNextLibrary",
            }
        );

        PrivateDependencyModuleNames.AddRange(
            new string[]
            {
                "Core",
                "NetworkNextLibrary",
                "Projects",
                "CoreUObject",
                "Engine",
                "Sockets",
                "OnlineSubsystem",
                "PacketHandler",
            }
        );

        if (UEBuildConfiguration.bBuildEditor == true)
        {
            PrivateDependencyModuleNames.Add("UnrealEd");
        }
    }
}
