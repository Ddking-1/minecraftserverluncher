# SPDX-License-Identifier: GPL-3.0-only
# The minecraft instance and launch subsystems are mutually dependent
# (MinecraftInstance::createLaunchTask assembles the launch steps, and the
# steps operate on MinecraftInstance), so they are built as one library.
add_library(launcher_minecraft STATIC
    MinecraftInstance.h
    MinecraftInstance.cpp
    PackProfile.h
    PackProfile.cpp
    ../launch/LaunchTask.h
    ../launch/LaunchTask.cpp
    ../launch/LaunchStep.h
    ../launch/LaunchStep.cpp
    ../launch/steps/ClaimAuthSession.h
    ../launch/steps/ClaimAuthSession.cpp
    ../launch/steps/ResolveLibraries.h
    ../launch/steps/ResolveLibraries.cpp
    ../launch/steps/ResolveAssets.h
    ../launch/steps/ResolveAssets.cpp
    ../launch/steps/ExtractNatives.h
    ../launch/steps/ExtractNatives.cpp
    ../launch/steps/VerifyJava.h
    ../launch/steps/VerifyJava.cpp
    ../launch/steps/LauncherPartLaunch.h
    ../launch/steps/LauncherPartLaunch.cpp
    ../launch/java/JavaInstall.h
    ../launch/java/JavaVersion.h
    ../launch/java/JavaVersion.cpp
)
target_include_directories(launcher_minecraft PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/..)
target_link_libraries(launcher_minecraft PUBLIC launcher_core Qt6::Core Qt6::Network)
