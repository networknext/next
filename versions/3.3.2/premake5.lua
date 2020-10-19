
solution "next"
	language "C++"
	configurations { "Debug", "Release" }
	rtti "Off"
	warnings "Extra"
	floatingpoint "Fast"
	platforms { "Static64" }
	flags { "FatalWarnings" }
	filter "options:static"
		staticruntime "On"
	filter "configurations:Debug"
		symbols "On"
	filter "configurations:Release"
		symbols "On"
		optimize "Speed"
		defines { "NDEBUG" }
	filter "platforms:*32"
		architecture "x86"
	filter "platforms:*64"
		architecture "x86_64"
	filter "platforms:*ARM"
		architecture "ARM"
	filter "system:not windows"
		targetdir "bin/"

project "next"
	kind "StaticLib"
	includedirs { "include" }
	links { "sodium" }
	files {
		"include/next.h",
		"src/next.cpp",
		"src/next_*.h",
		"src/next_*.cpp",
	}
		
project "client"
	kind "ConsoleApp"
	includedirs { "include" }
	links { "next", "sodium" }
	files { "client.cpp" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "server"
	kind "ConsoleApp"
	includedirs { "include" }
	links { "next", "sodium" }
	files { "server.cpp" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
