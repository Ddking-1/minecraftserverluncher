# SPDX-License-Identifier: GPL-3.0-only
# Project-owned launcher core. Root build configuration for the C++ launcher.
cmake_minimum_required(VERSION 3.16)
project(MinecraftLauncher LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

# The approved architecture (docs/DECISIONS.md DECISION-001/003) requires Qt.
# Qt 6 is NOT installed in the current build environment; configure will fail
# at find_package below until Qt is provided (see launcher/README.md).
find_package(Qt6 6.2 REQUIRED COMPONENTS Core Gui Widgets Network)

add_subdirectory(core)
add_subdirectory(minecraft)
add_subdirectory(auth)

# Minimal launcher entry point (UI not yet implemented).
add_executable(MinecraftLauncher main.cpp)
target_link_libraries(MinecraftLauncher
    PRIVATE
        launcher_core
        launcher_minecraft
        launcher_auth
        Qt6::Core
        Qt6::Gui
        Qt6::Widgets
        Qt6::Network
)

if(WIN32)
    target_compile_definitions(MinecraftLauncher PRIVATE _CRT_SECURE_NO_WARNINGS)
endif()
