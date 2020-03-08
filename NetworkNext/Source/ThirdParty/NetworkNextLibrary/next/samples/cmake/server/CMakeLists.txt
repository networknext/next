cmake_minimum_required (VERSION 3.13)

add_executable (server "${CMAKE_CURRENT_SOURCE_DIR}/../../../server.cpp")

target_include_directories(server PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}/../../../include")
target_link_directories(server PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}/../../../lib/${NEXT_LIBRARY_ARCH}/${CMAKE_BUILD_TYPE}")
target_link_libraries(server PUBLIC "next${NEXT_LIBRARY_LINK_EXT}")
add_custom_command(TARGET server POST_BUILD COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_SOURCE_DIR}/../../../lib/${NEXT_LIBRARY_ARCH}/${CMAKE_BUILD_TYPE}/${NEXT_LIBRARY_PREFIX}next${NEXT_LIBRARY_LINK_EXT}${NEXT_LIBRARY_EXT}" $<TARGET_FILE_DIR:server>)
