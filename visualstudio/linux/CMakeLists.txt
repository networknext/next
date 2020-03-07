cmake_minimum_required (VERSION 3.8)

# Common setup (no trailing slash on outputDirRoot)
project ("next_linux")
set(outputDirRoot /mnt/c/UE4_Projects/NetworkNextTest/Plugins/NetworkNext/Source/ThirdParty/NetworkNextLibrary/next/lib/Linux)

# libsodium common
file (GLOB_RECURSE SODIUM_SOURCE_FILES ${CMAKE_CURRENT_SOURCE_DIR}/src/sodium/src/libsodium/*.c)
file (GLOB_RECURSE SODIUM_HEADER_FILES ${CMAKE_CURRENT_SOURCE_DIR}/src/sodium/src/libsodium/*.h)

# libsodium 32 bit
add_library (sodium_static_x86 STATIC ${SODIUM_SOURCE_FILES} ${SODIUM_HEADER_FILES})
set_property(TARGET sodium_static_x86 PROPERTY POSITION_INDEPENDENT_CODE ON)
target_include_directories (sodium_static_x86 PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/src/sodium/src/libsodium/include/sodium/)
target_compile_definitions (sodium_static_x86 PUBLIC "SODIUM_STATIC" "SODIUM_EXPORT=" "CONFIGURED=1")

set_target_properties(sodium_static_x86
    PROPERTIES
	COMPILE_FLAGS "-m32"
	LINK_FLAGS "-m32"
    ARCHIVE_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86/Static-Release
	ARCHIVE_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86/Static-Debug
)

# libsodium 64 bit
add_library (sodium_static_x64 STATIC ${SODIUM_SOURCE_FILES} ${SODIUM_HEADER_FILES})
set_property(TARGET sodium_static_x64 PROPERTY POSITION_INDEPENDENT_CODE ON)
target_include_directories (sodium_static_x64 PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/src/sodium/src/libsodium/include/sodium/)
target_compile_definitions (sodium_static_x64 PUBLIC "SODIUM_STATIC" "SODIUM_EXPORT=" "CONFIGURED=1")

set_target_properties(sodium_static_x64
    PROPERTIES
	COMPILE_FLAGS "-m64"
	LINK_FLAGS "-m64"
    ARCHIVE_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86_64/Static-Release
	ARCHIVE_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86_64/Static-Debug
)


# next common
set (NEXT_SOURCE_FILES ${CMAKE_CURRENT_SOURCE_DIR}/src/next.cpp ${CMAKE_CURRENT_SOURCE_DIR}/src/next_linux.cpp)
set (NEXT_HEADER_FILES ${CMAKE_CURRENT_SOURCE_DIR}/src/next.cpp ${CMAKE_CURRENT_SOURCE_DIR}/src/next_linux.h)

# next 32 bit static
add_library (next_static_x86 STATIC ${NEXT_SOURCE_FILES} ${NEXT_HEADER_FILES})
target_compile_definitions (next_static_x86 PUBLIC $<$<CONFIG:Release>:"NDEBUG"> "NEXT_EXPORT=" "SODIUM_STATIC")
target_link_libraries (next_static_x86 sodium_static_x86)

set_target_properties(next_static_x86
    PROPERTIES
	COMPILE_FLAGS "-m32" LINK_FLAGS "-m32"
    ARCHIVE_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86/Static-Release
	ARCHIVE_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86/Static-Debug
)

# next 64 bit static
add_library (next_static_x64 STATIC ${NEXT_SOURCE_FILES} ${NEXT_HEADER_FILES})
target_compile_definitions (next_static_x64 PUBLIC $<$<CONFIG:Release>:"NDEBUG"> "NEXT_EXPORT=" "SODIUM_STATIC")
target_link_libraries (next_static_x64 sodium_static_x64)

set_target_properties(next_static_x64
    PROPERTIES
	COMPILE_FLAGS "-m64" LINK_FLAGS "-m64"
    ARCHIVE_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86_64/Static-Release
	ARCHIVE_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86_64/Static-Debug
)

# next 32 bit shared
add_library (next_shared_x86 SHARED ${NEXT_SOURCE_FILES} ${NEXT_HEADER_FILES})
target_compile_definitions (next_shared_x86 PUBLIC $<$<CONFIG:Release>:"NDEBUG"> "NEXT_SHARED" "NEXT_EXPORT=" "SODIUM_STATIC")
target_link_libraries (next_shared_x86 sodium_static_x86)

set_target_properties(next_shared_x86
    PROPERTIES
	COMPILE_FLAGS "-m32" LINK_FLAGS "-m32"
    LIBRARY_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86/Shared-Release
	LIBRARY_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86/Shared-Debug
)

# next 64 bit shared
add_library (next_shared_x64 SHARED ${NEXT_SOURCE_FILES} ${NEXT_HEADER_FILES})
target_compile_definitions (next_shared_x64 PUBLIC $<$<CONFIG:Release>:"NDEBUG"> "NEXT_SHARED" "NEXT_EXPORT=" "SODIUM_STATIC")
target_link_libraries (next_shared_x64 sodium_static_x64)

set_target_properties(next_shared_x64
    PROPERTIES
	COMPILE_FLAGS "-m64" LINK_FLAGS "-m64"
    LIBRARY_OUTPUT_DIRECTORY_RELEASE ${outputDirRoot}-x86_64/Shared-Release
	LIBRARY_OUTPUT_DIRECTORY_DEBUG ${outputDirRoot}-x86_64/Shared-Debug
)