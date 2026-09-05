# SPDX-License-Identifier: GPL-3.0-only
add_library(launcher_auth STATIC
    BackendAuthClient.h
    BackendAuthClient.cpp
    SessionManager.h
    SessionManager.cpp
    TokenStorage.h
    TokenStorage.cpp
)
target_include_directories(launcher_auth PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/..)
target_link_libraries(launcher_auth PUBLIC launcher_core Qt6::Core Qt6::Network)
