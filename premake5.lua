solution "enet"
	configurations { "Debug", "Release" }
	defines { "ENET_NETWORK_NEXT=1" }
	platforms { "portable", "x64" }

project "next"
	kind "StaticLib"
	links { "sodium" }
	files {
		"next/next.h",
		"next/next.cpp",
		"next/next_*.h",
		"next/next_*.cpp",
	}
	includedirs { "next", "sodium" }
	filter "system:windows"
		linkoptions { "/ignore:4221" }
		disablewarnings { "4324" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "enet"
	kind "StaticLib"
	language "C"
	links { "next" }
	files { "enet/*.c" }
	includedirs { "enet", "next", "sodium" }
	configuration "Debug"
		targetsuffix "d"
		defines({ "DEBUG" })
		symbols "On"		
	configuration "Release"
		defines({ "NDEBUG" })
		optimize "On"
	configuration { "Debug", "x64" }
		targetsuffix "64d"
	configuration { "Release", "x64" }
		targetsuffix "64"

project "sodium"
	kind "StaticLib"
	includedirs { "sodium" }
	files {
		"sodium/**.c",
		"sodium/**.h",
	}
  	filter { "system:not windows", "platforms:*x64 or *avx or *avx2" }
		files {
			"sodium/**.S"
		}
	filter "platforms:*x86"
		architecture "x86"
		defines { "NEXT_X86=1" }
	filter "platforms:*x64"
		architecture "x86_64"
		defines { "NEXT_X64=1" }
	filter "platforms:*avx"
		architecture "x86_64"
		vectorextensions "AVX"
		defines { "NEXT_X64=1", "NEXT_AVX=1" }
	filter "platforms:*avx2"
		architecture "x86_64"
		vectorextensions "AVX2"
		defines { "NEXT_X64=1", "NEXT_AVX=1", "NEXT_AVX2=1" }
	filter "system:windows"
		disablewarnings { "4221", "4244", "4715", "4197", "4146", "4324", "4456", "4100", "4459", "4245" }
		linkoptions { "/ignore:4221" }
	configuration { "gmake" }
  		buildoptions { "-Wno-unused-parameter", "-Wno-unused-function", "-Wno-unknown-pragmas", "-Wno-unused-variable", "-Wno-type-limits" }

project "client"
	kind "ConsoleApp"
	links { "enet", "next", "sodium" }
	files { "client.cpp" }
	includedirs { "enet", "next" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "server"
	kind "ConsoleApp"
	links { "enet", "next", "sodium" }
	files { "server.cpp" }
	includedirs { "enet", "next" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
