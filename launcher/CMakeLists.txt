# launcher/CMakeLists.txt

# Set minimum required version of CMake
cmake_minimum_required(VERSION 3.10)

# Set project name and version
project(MinecraftLauncher VERSION 1.0)

# Set C++ standard to C++17
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Add the executable target
add_executable(${PROJECT_NAME} 
    main.cpp
    core/HWIDExtractor.cpp
)

# Include directories
target_include_directories(${PROJECT_NAME} PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

# Link libraries (Windows specific)
if(WIN32)
    target_link_libraries(${PROJECT_NAME} 
        ws2_32
        wmiutils
        iphlpapi
        Crypt32.lib  # For cryptographic functions
    )
endif()

# Enable C++ exceptions and RTTI
set_property(TARGET ${PROJECT_NAME} PROPERTY MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:d>")

# Add compiler options for Windows (MSVC)
if(MSVC)
    add_compile_options(/W4 /WX)
else()
    # For GCC/Clang, you might want to add extra warnings
    add_compile_options(-Wall -Wextra)
endif()

# Enable C++17 features 
add_compile_definitions(_CRT_SECURE_NO_WARNINGS)

# Add the header files for IDE integration
set(HEADERS
    core/HWIDExtractor.h
)

# Group header files in Visual Studio if using VS generator
if(CMAKE_GENERATOR MATCHES "Visual Studio")
    set_property(GLOBAL PROPERTY VS_GLOBAL_KEYWORD "Win32")
endif()