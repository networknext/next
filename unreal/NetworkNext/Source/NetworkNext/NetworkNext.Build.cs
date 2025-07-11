/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

using UnrealBuildTool;

public class NetworkNext : ModuleRules
{
    public NetworkNext(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = ModuleRules.PCHUsageMode.UseExplicitOrSharedPCHs;

        PublicDefinitions.Add("NETWORK_NEXT_UNREAL_ENGINE=1");

        PublicDependencyModuleNames.AddRange(
            new string[] {
                "OnlineSubsystemUtils",
            }
        );

        PrivateDependencyModuleNames.AddRange(
            new string[]
            {
                "Core",
                "Projects",
                "CoreUObject",
                "Engine",
                "Sockets",
                "OnlineSubsystem",
                "PacketHandler",
            }
        );

        PrivateIncludePaths.AddRange(new string[] { "NetworkNext/Private/include", "NetworkNext/Private/sodium" });

        if (Target.Platform == UnrealTargetPlatform.Mac)
        {
            PublicFrameworks.Add("SystemConfiguration");
        } 
        else if (Target.Platform.ToString() == "PS4")
        {
            string SDKDir = System.Environment.GetEnvironmentVariable("SCE_ORBIS_SDK_DIR");
            string LibDir = System.IO.Path.Combine(SDKDir, "target", "lib");
            PublicAdditionalLibraries.Add(System.IO.Path.Combine(LibDir, "libSceSecure.a"));
        }
    }
}
