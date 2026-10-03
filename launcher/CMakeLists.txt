cmake_minimum_required(VERSION 3.16)
project(MinecraftLauncher)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 REQUIRED COMPONENTS Core Network)

add_executable(MinecraftLauncher
    main.cpp
    core/BaseInstance.cpp
    core/InstanceList.cpp
    core/InstancePersistence.cpp
    core/HWIDExtractor.cpp
    game/JavaDetector.cpp
    game/MinecraftLauncher.cpp
    minecraft/MinecraftInstance.cpp
    minecraft/PackProfile.cpp
    launch/LaunchTask.cpp
    launch/LaunchStep.cpp
    launch/java/JavaVersion.cpp
    launch/steps/ClaimAuthSession.cpp
    launch/steps/ExtractNatives.cpp
    launch/steps/LauncherPartLaunch.cpp
    launch/steps/ResolveAssets.cpp
    launch/steps/ResolveLibraries.cpp
    launch/steps/VerifyJava.cpp
)

target_include_directories(MinecraftLauncher PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(MinecraftLauncher PRIVATE Qt6::Core Qt6::Network)

if(WIN32)
    target_link_libraries(MinecraftLauncher PRIVATE bcrypt crypt32 iphlpapi wbemuuid ole32 oleaut32)
endif()
