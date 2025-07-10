cmake_minimum_required(VERSION 3.16)
project(GoRDP-GUI VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find Qt6 components
find_package(Qt6 REQUIRED COMPONENTS 
    Core 
    Widgets 
    Network 
    WebSockets
    Gui
    Charts
)

# Enable Qt's automatic MOC, UIC, and RCC
qt_standard_project_setup()

# Set output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# Source files
set(SOURCES
    src/main.cpp
    src/mainwindow/mainwindow.cpp
    src/connection/connection_dialog.cpp
    src/display/rdp_display.cpp
    src/input/input_handler.cpp
    src/settings/settings_dialog.cpp
    src/performance/performance_dialog.cpp
    src/history/history_dialog.cpp
    src/history/connection_history.cpp
    src/favorites/favorites_dialog.cpp
    src/favorites/favorites_manager.cpp
    src/plugins/plugin_manager.cpp
    src/virtualchannels/virtual_channel_dialog.cpp
    src/multimonitor/monitor_dialog.cpp
    src/multimonitor/monitor_layout_preview_widget.cpp
    src/utils/gordp_bridge.cpp
)

# Header files
set(HEADERS
    src/mainwindow/mainwindow.h
    src/connection/connection_dialog.h
    src/display/rdp_display.h
    src/input/input_handler.h
    src/settings/settings_dialog.h
    src/performance/performance_dialog.h
    src/history/history_dialog.h
    src/history/connection_history.h
    src/favorites/favorites_dialog.h
    src/favorites/favorites_manager.h
    src/plugins/plugin_manager.h
    src/virtualchannels/virtual_channel_dialog.h
    src/multimonitor/monitor_dialog.h
    src/multimonitor/monitor_layout_preview_widget.h
    src/utils/gordp_bridge.h
)

# UI files
set(UI_FILES
    src/mainwindow/mainwindow.ui
    src/connection/connection_dialog.ui
    src/settings/settings_dialog.ui
    src/performance/performance_dialog.ui
    src/history/history_dialog.ui
    src/favorites/favorites_dialog.ui
    src/plugins/plugin_manager.ui
    src/virtualchannels/virtual_channel_dialog.ui
    src/multimonitor/monitor_dialog.ui
)

# Resource files
set(RESOURCES
    resources/resources.qrc
)

# Create executable
qt_add_executable(gordp-gui
    ${SOURCES}
    ${HEADERS}
    ${UI_FILES}
    ${RESOURCES}
)

# Link Qt libraries
target_link_libraries(gordp-gui PRIVATE
    Qt6::Core
    Qt6::Widgets
    Qt6::Network
    Qt6::WebSockets
    Qt6::Gui
    Qt6::Charts
)

# Set include directories
target_include_directories(gordp-gui PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# Platform-specific settings
if(WIN32)
    set_target_properties(gordp-gui PROPERTIES
        WIN32_EXECUTABLE TRUE
    )
endif()

# Install rules
install(TARGETS gordp-gui
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

# Copy Qt runtime dependencies (for deployment)
if(WIN32)
    qt_deploy_runtime_dependencies(EXECUTABLE gordp-gui)
endif()

# Print configuration info
message(STATUS "GoRDP GUI Configuration:")
message(STATUS "  Qt Version: ${Qt6_VERSION}")
message(STATUS "  C++ Standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "  Build Type: ${CMAKE_BUILD_TYPE}")
message(STATUS "  Output Directory: ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}") 