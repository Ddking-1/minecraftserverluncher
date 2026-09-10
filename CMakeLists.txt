# SPDX-License-Identifier: GPL-3.0-only
cmake_minimum_required(VERSION 3.16)
project(MinecraftLauncher)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(CMAKE_AUTOMOC ON)

# Find Qt5 or Qt6 (try both)
find_package(Qt5 COMPONENTS Core Widgets Network REQUIRED)
if(NOT QT_FOUND)
    find_package(Qt6 COMPONENTS Core Widgets Network REQUIRED)
endif()

add_subdirectory(launcher)

add_subdirectory(launcher)