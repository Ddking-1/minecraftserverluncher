cmake_minimum_required(VERSION 3.16)
project(MinecraftLauncher)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 REQUIRED COMPONENTS Core Network)

add_subdirectory(auth)
add_subdirectory(core)
add_subdirectory(minecraft)

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
target_link_libraries(MinecraftLauncher PRIVATE launcher_auth launcher_minecraft launcher_core Qt6::Core Qt6::Network)

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

add_executable(test_editor_config tests/test_editor_config.cpp core/EditorConfig.cpp core/InstancePersistence.cpp)
target_include_directories(test_editor_config PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_editor_config PRIVATE Qt6::Core)
add_test(NAME editor_config COMMAND test_editor_config)

add_executable(test_skin_client tests/test_skin_client.cpp core/SkinClient.cpp auth/BackendAuthClient.cpp)
target_include_directories(test_skin_client PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_skin_client PRIVATE Qt6::Core Qt6::Network)
add_test(NAME skin_client COMMAND test_skin_client)

add_executable(test_download_asset tests/test_download_asset.cpp core/DownloadAsset.cpp core/FileIntegrity.cpp)
target_include_directories(test_download_asset PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_download_asset PRIVATE Qt6::Core Qt6::Network)
add_test(NAME download_asset COMMAND test_download_asset)

add_executable(test_download_queue tests/test_download_queue.cpp core/DownloadQueue.cpp core/DownloadAsset.cpp core/FileIntegrity.cpp)
target_include_directories(test_download_queue PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_download_queue PRIVATE Qt6::Core Qt6::Network)
add_test(NAME download_queue COMMAND test_download_queue)

add_executable(test_java_detector tests/test_java_detector.cpp game/JavaDetector.cpp)
target_include_directories(test_java_detector PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_java_detector PRIVATE Qt6::Core)
add_test(NAME java_detector COMMAND test_java_detector)

add_executable(test_game_paths tests/test_game_paths.cpp core/GamePaths.cpp)
target_include_directories(test_game_paths PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_game_paths PRIVATE Qt6::Core)
add_test(NAME game_paths COMMAND test_game_paths)

add_executable(test_game_launcher tests/test_game_launcher.cpp core/GameLauncher.cpp game/JavaDetector.cpp)
target_include_directories(test_game_launcher PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(test_game_launcher PRIVATE Qt6::Core)
add_test(NAME game_launcher COMMAND test_game_launcher)
