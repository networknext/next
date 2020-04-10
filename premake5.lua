if os.ishost("windows") then
	local win10SDK = os.getWindowsRegistry("HKLM:SOFTWARE\\Wow6432Node\\Microsoft\\Microsoft SDKs\\Windows\\v10.0\\ProductVersion")
	-- It needs an extra ".0" to be recognized by VS
	if win10SDK ~= nil then systemversion(win10SDK .. ".0") end
end

newoption {
	trigger = "xp",
	description = "Windows XP"
}

newoption {
	trigger = "static",
	description = "Link against static runtime",
}

solution "next"
	language "C++"
	configurations { "Debug", "Release" }
	targetdir "bin/"
	rtti "Off"
	warnings "Extra"
	floatingpoint "Fast"
	platforms { "Static64", "Shared64", "StaticArm", "SharedARM", "Shared32", "Static32" }
	flags { "FatalWarnings" }
	filter "options:static"
		staticruntime "On"
	filter { "action:gmake*", "toolset:gcc" }
		disablewarnings{ "class-memaccess" }
	filter "options:xp"
		defines { "WIN32=1", "_WIN32_WINNT=0x0502", "WINVER=0x0502" }
		disablewarnings{ "4127" }
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
	filter "platforms:Shared*"
		defines { "NEXT_SHARED" }

project "next"
	defines { "NEXT_EXPORT", "SODIUM_STATIC" }
	links { "sodium" }
	files {
		"include/next.h",
		"source/next.cpp",
		"source/next_*.h",
		"source/next_*.cpp",
	}
	includedirs { "include" }
	filter "system:not windows"
		links { "pthread" }
		targetname "next-%{cfg.platform}-%{cfg.buildcfg}"
	filter "system:windows"
		linkoptions { "/ignore:4221" }
		disablewarnings { "4324" }
		includedirs { "src/sodium/src/libsodium/include/" }
	filter "platforms:Shared*"
		kind "SharedLib"
	filter "platforms:Static*"
		kind "StaticLib"
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

if os.ishost("windows") then
project "sodium"
	kind "StaticLib"
	defines { "SODIUM_STATIC", "SODIUM_EXPORT=", "CONFIGURED=1" }
	includedirs { "src/sodium/src/libsodium/include/sodium/" }
	linkoptions { "/ignore:4221" }
	files {
		"sodium/**.c",
		"sodium/**.h"
	}
	disablewarnings { "4244", "4715", "4197", "4146", "4324", "4456", "4100", "4459", "4245" }
	linkoptions { "/ignore:4221" }
	filter "options:not internal"
		vectorextensions "AVX"
	filter { "action:vs2010"}
		defines { "inline=__inline;NATIVE_LITTLE_ENDIAN;_CRT_SECURE_NO_WARNINGS;" }
	configuration { "gmake" }
  		buildoptions { "-Wno-unused-parameter", "-Wno-unused-function", "-Wno-unknown-pragmas", "-Wno-unused-variable" }
end

project "client"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "client.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
		includedirs { "sodium/include/" }
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
		includedirs { "sodium/include/" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "test"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "test.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
		includedirs { "sodium/include/" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "soak"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "soak.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
		includedirs { "sodium/include/" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
