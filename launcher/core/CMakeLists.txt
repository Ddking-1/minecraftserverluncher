# SPDX-License-Identifier: GPL-3.0-only
add_library(launcher_core STATIC
    BaseInstance.h
    BaseInstance.cpp
    InstanceList.h
    InstanceList.cpp
    InstancePersistence.h
    InstancePersistence.cpp
    InstanceTypes.h
    SystemMemory.h
    SystemMemory.cpp
    GameExitDiagnostics.h
    GameExitDiagnostics.cpp
    FriendlyError.h
    FriendlyError.cpp
)
target_include_directories(launcher_core PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/..)
target_link_libraries(launcher_core PUBLIC Qt6::Core)
