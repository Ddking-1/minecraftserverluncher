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
    core/SkinModel.cpp
    core/SkinCache.cpp
    core/SettingsStore.cpp
    core/RotatingFileLogger.cpp
    core/FileIntegrity.cpp
    core/HWIDExtractor.cpp
    game/JavaDetector.cpp
    game/MinecraftLauncher.cpp
    minecraft/MinecraftInstance.cpp
    minecraft/PackProfile.cpp
    minecraft/ModScanner.cpp
    launch/LaunchTask.cpp
    launch/LaunchStep.cpp
    launch/java/JavaVersion.cpp
    launch/steps/ClaimAuthSession.cpp
    launch/steps/ExtractNatives.cpp
    launch/steps/LauncherPartLaunch.cpp
    launch/steps/ResolveAssets.cpp
    launch/steps/ResolveLibraries.cpp
    launch/steps/VerifyJava.cpp
    updates/UpdateManifest.cpp
    updates/UpdateChecker.cpp
)

target_include_directories(MinecraftLauncher PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(MinecraftLauncher PRIVATE Qt6::Core Qt6::Network)

if(WIN32)
    target_link_libraries(MinecraftLauncher PRIVATE bcrypt crypt32 iphlpapi wbemuuid ole32 oleaut32)
endif()

enable_testing()

add_executable(test_hwid tests/test_hwid.cpp core/HWIDExtractor.cpp)
target_include_directories(test_hwid PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
if(WIN32)
    target_link_libraries(test_hwid PRIVATE bcrypt iphlpapi)
endif()
add_test(NAME hwid COMMAND test_hwid)

add_executable(test_settings tests/test_settings.cpp core/InstancePersistence.cpp)
target_include_directories(test_settings PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_settings PRIVATE Qt6::Core)
add_test(NAME settings COMMAND test_settings)

add_executable(test_update_manifest tests/test_update_manifest.cpp updates/UpdateManifest.cpp)
target_include_directories(test_update_manifest PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_update_manifest PRIVATE Qt6::Core)
add_test(NAME update_manifest COMMAND test_update_manifest)

add_executable(test_skin_model tests/test_skin_model.cpp core/SkinModel.cpp)
target_include_directories(test_skin_model PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_skin_model PRIVATE Qt6::Core)
add_test(NAME skin_model COMMAND test_skin_model)

add_executable(test_backend_contract tests/test_backend_contract.cpp auth/BackendAuthClient.cpp)
target_include_directories(test_backend_contract PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_backend_contract PRIVATE Qt6::Core Qt6::Network)
add_test(NAME backend_contract COMMAND test_backend_contract)

add_executable(test_settings_store tests/test_settings_store.cpp core/SettingsStore.cpp)
target_include_directories(test_settings_store PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_settings_store PRIVATE Qt6::Core)
add_test(NAME settings_store COMMAND test_settings_store)

add_executable(test_mod_scanner tests/test_mod_scanner.cpp minecraft/ModScanner.cpp)
target_include_directories(test_mod_scanner PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_mod_scanner PRIVATE Qt6::Core)
add_test(NAME mod_scanner COMMAND test_mod_scanner)

add_executable(test_skin_cache tests/test_skin_cache.cpp core/SkinCache.cpp core/SkinModel.cpp)
target_include_directories(test_skin_cache PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_skin_cache PRIVATE Qt6::Core)
add_test(NAME skin_cache COMMAND test_skin_cache)

add_executable(test_rotating_logger tests/test_rotating_logger.cpp core/RotatingFileLogger.cpp)
target_include_directories(test_rotating_logger PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_rotating_logger PRIVATE Qt6::Core)
add_test(NAME rotating_logger COMMAND test_rotating_logger)

add_executable(test_update_checker tests/test_update_checker.cpp updates/UpdateChecker.cpp)
target_include_directories(test_update_checker PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_update_checker PRIVATE Qt6::Core)
add_test(NAME update_checker COMMAND test_update_checker)

add_executable(test_file_integrity tests/test_file_integrity.cpp core/FileIntegrity.cpp)
target_include_directories(test_file_integrity PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_file_integrity PRIVATE Qt6::Core)
add_test(NAME file_integrity COMMAND test_file_integrity)

add_executable(test_token_storage tests/test_token_storage.cpp auth/TokenStorage.cpp)
target_include_directories(test_token_storage PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_token_storage PRIVATE Qt6::Core)
if(WIN32)
    target_link_libraries(test_token_storage PRIVATE crypt32)
endif()
add_test(NAME token_storage COMMAND test_token_storage)

add_executable(test_resolve_libraries
    tests/test_resolve_libraries.cpp
    auth/TokenStorage.cpp
    auth/SessionManager.cpp
    auth/BackendAuthClient.cpp
    launch/LaunchTask.cpp
    launch/LaunchStep.cpp
    launch/java/JavaVersion.cpp
    launch/steps/ClaimAuthSession.cpp
    launch/steps/ExtractNatives.cpp
    launch/steps/LauncherPartLaunch.cpp
    launch/steps/ResolveAssets.cpp
    launch/steps/ResolveLibraries.cpp
    launch/steps/VerifyJava.cpp
    minecraft/MinecraftInstance.cpp
    minecraft/PackProfile.cpp
    core/BaseInstance.cpp
    core/InstancePersistence.cpp
)
target_include_directories(test_resolve_libraries PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_resolve_libraries PRIVATE Qt6::Core Qt6::Network)
add_test(NAME resolve_libraries COMMAND test_resolve_libraries)

add_executable(test_resolve_assets
    tests/test_resolve_assets.cpp
    auth/TokenStorage.cpp
    auth/SessionManager.cpp
    auth/BackendAuthClient.cpp
    launch/LaunchTask.cpp
    launch/LaunchStep.cpp
    launch/java/JavaVersion.cpp
    launch/steps/ClaimAuthSession.cpp
    launch/steps/ExtractNatives.cpp
    launch/steps/LauncherPartLaunch.cpp
    launch/steps/ResolveAssets.cpp
    launch/steps/ResolveLibraries.cpp
    launch/steps/VerifyJava.cpp
    minecraft/MinecraftInstance.cpp
    minecraft/PackProfile.cpp
    core/BaseInstance.cpp
    core/InstancePersistence.cpp
)
target_include_directories(test_resolve_assets PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_resolve_assets PRIVATE Qt6::Core Qt6::Network)
add_test(NAME resolve_assets COMMAND test_resolve_assets)
