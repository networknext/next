
solution "next"
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
		symbols "On"
		optimize "Speed"
		defines { "NDEBUG" }
	filter "system:windows"
		architecture "x86_64"
	filter "system:windows"
		location ("visualstudio")

project "next"
	kind "StaticLib"
	defines { "NEXT_EXPORT", "SODIUM_STATIC" }
	links { "sodium" }
	files {
		"include/next.h",
		"source/next.cpp",
		"source/next_*.h",
		"source/next_*.cpp",
	}
	includedirs { "include", "sodium/include/" }
	filter "system:windows"
		linkoptions { "/ignore:4221" }
		disablewarnings { "4324" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "sodium"
	kind "StaticLib"
	defines { "SODIUM_STATIC", "SODIUM_EXPORT=", "CONFIGURED=1" }
	includedirs { "sodium/include/sodium" }
	files {
		"sodium/**.c",
		"sodium/**.h"
	}
	filter "system:windows"
		disablewarnings { "4221", "4244", "4715", "4197", "4146", "4324", "4456", "4100", "4459", "4245" }
		linkoptions { "/ignore:4221" }
	filter { "action:vs2010"}
		defines { "inline=__inline;NATIVE_LITTLE_ENDIAN;_CRT_SECURE_NO_WARNINGS;" }
	configuration { "gmake" }
  		buildoptions { "-Wno-unused-parameter", "-Wno-unused-function", "-Wno-unknown-pragmas", "-Wno-unused-variable", "-Wno-type-limits" }

project "client"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "client.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "server"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "server.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
