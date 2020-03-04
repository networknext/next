solution "next"
    kind "ConsoleApp"
    language "C++"
    configurations { "Release", "Debug" }
    targetdir "bin/"
    rtti "Off"
    warnings "Extra"
    floatingpoint "Fast"
    configuration "Debug"
        symbols "On"
    configuration "Release"
        symbols "On"
        optimize "Speed"
        defines { "NDEBUG" }

project "relay"
    defines { "_GNU_SOURCE" }
    links { "sodium", "pthread", "curl" }
    includedirs { "deps" }
    files {
        "deps/miniz.cpp",
        "relay.cpp",
        "relay.h",
        "relay_internal.h",
        "relay_internal.cpp",
        "relay_unix.h",
        "relay_unix.cpp"
    }
    buildoptions { "-std=c++11 -fpermissive" }

newaction
{
    trigger     = "clean",

    description = "Clean all build files and output",

    execute = function ()

        files_to_delete =
        {
            "Makefile",
            "*.make",
            "*.zip",
            "*.tar.gz",
            "*.db",
            "*.opendb",
            "*.vcproj",
            "*.vcxproj",
            "*.vcxproj.user",
            "*.vcxproj.filters",
            "*.sln",
            "*.xcodeproj",
            "*.xcworkspace",
            "master/*.go",
            "*.txt",
        }

        directories_to_delete =
        {
            "obj",
            "ipch",
            "bin",
            ".vs",
            "Debug",
            "Release",
            "release",
            "docs",
            "xml",
        }

        for i,v in ipairs( directories_to_delete ) do
          os.rmdir( v )
        end

        if not os.ishost "windows" then
            os.execute "find . -name .DS_Store -delete"
            for i,v in ipairs( files_to_delete ) do
              os.execute( "rm -f " .. v )
            end
        else
            for i,v in ipairs( files_to_delete ) do
              os.execute( "del /F /Q  " .. v )
            end
        end

    end
}
