
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
	includedirs { "include", "sodium" }
	filter "system:windows"
		linkoptions { "/ignore:4221" }
		disablewarnings { "4324" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "sodium"
	kind "StaticLib"
	defines { "SODIUM_STATIC", "SODIUM_EXPORT=", "CONFIGURED=1", "HAVE_TI_MODE=1", "HAVE_AVX_ASM=1" }
	includedirs { "sodium" }
	files {
		"sodium/**.c",
		"sodium/**.h",
		"sodium/**.S"
	}
	filter "system:windows"
		disablewarnings { "4221", "4244", "4715", "4197", "4146", "4324", "4456", "4100", "4459", "4245" }
		linkoptions { "/ignore:4221" }
	filter { "action:vs2010"}
		defines { "inline=__inline;NATIVE_LITTLE_ENDIAN;_CRT_SECURE_NO_WARNINGS;" }
	configuration { "gmake" }
  		buildoptions { "-Wno-unused-parameter", "-Wno-unused-function", "-Wno-unknown-pragmas", "-Wno-unused-variable", "-Wno-type-limits" }

project "test"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "test.cpp" }
	includedirs { "include" }
	filter "system:windows"
		disablewarnings { "4324" }
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
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "simple_client"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/simple_client.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "simple_server"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/simple_server.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "upgraded_client"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/upgraded_client.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "upgraded_server"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/upgraded_server.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "complex_client"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/complex_client.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }

project "complex_server"
	kind "ConsoleApp"
	links { "next", "sodium" }
	files { "examples/complex_server.cpp" }
	includedirs { "include" }
	defines { "SODIUM_STATIC" }
	filter "system:windows"
		disablewarnings { "4324" }
	filter "system:not windows"
		links { "pthread" }
	filter "system:macosx"
		linkoptions { "-framework SystemConfiguration -framework CoreFoundation" }
