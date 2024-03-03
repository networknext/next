
solution "relay"
	configurations { "Debug", "Release" }
	targetdir "bin/"
	rtti "Off"
	warnings "Extra"
	floatingpoint "Fast"
	flags { "FatalWarnings" }
	filter "configurations:Debug"
		symbols "On"
		defines { "_DEBUG" }
	filter "configurations:Release"
		optimize "Speed"
		defines { "NDEBUG" }
		editandcontinue "Off"

project "relay"
	kind "ConsoleApp"
	links { "sodium", "pthread" ,"curl" }
	files { "*.cpp", "*.h" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "debug"
	defines { "RELAY_DEBUG" }
	kind "ConsoleApp"
	links { "sodium", "pthread" ,"curl" }
	files { "*.cpp", "*.h" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
