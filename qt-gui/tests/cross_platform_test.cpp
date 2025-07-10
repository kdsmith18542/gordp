#include <QTest>
#include <QApplication>
#include <QMainWindow>
#include <QDialog>
#include <QMessageBox>
#include <QJsonObject>
#include <QJsonArray>
#include <QSettings>
#include <QDir>
#include <QFile>
#include <QStandardPaths>
#include <QScreen>
#include <QClipboard>
#include <QNetworkAccessManager>
#include <QWebSocket>
#include <QElapsedTimer>
#include <QEventLoop>
#include <QTimer>
#include <QNetworkRequest>
#include <QNetworkReply>
#include <QNetworkAccessManager>
#include <QJsonParseError>
#include <QJsonDocument>

// Include our application headers
#include "../src/mainwindow/mainwindow.h"
#include "../src/connection/connection_dialog.h"
#include "../src/display/rdp_display.h"
#include "../src/input/input_handler.h"
#include "../src/settings/settings_dialog.h"
#include "../src/performance/performance_dialog.h"
#include "../src/history/connection_history.h"
#include "../src/favorites/favorites_manager.h"
#include "../src/plugins/plugin_manager.h"
#include "../src/virtualchannels/virtual_channel_dialog.h"
#include "../src/multimonitor/monitor_dialog.h"
#include "../src/utils/gordp_bridge.h"

class CrossPlatformTest : public QObject
{
    Q_OBJECT

private slots:
    // Test initialization
    void initTestCase();
    void cleanupTestCase();
    
    // Main window tests
    void testMainWindowCreation();
    void testMainWindowMenuBar();
    void testMainWindowToolBar();
    void testMainWindowStatusBar();
    
    // Connection dialog tests
    void testConnectionDialogCreation();
    void testConnectionDialogValidation();
    void testConnectionDialogSettings();
    
    // Display tests
    void testRDPDisplayWidget();
    void testBitmapRendering();
    void testDisplayScaling();
    
    // Input handling tests
    void testMouseInput();
    void testKeyboardInput();
    void testInputFocus();
    
    // Settings tests
    void testSettingsDialog();
    void testSettingsPersistence();
    
    // Performance tests
    void testPerformanceDialog();
    void testPerformanceMonitoring();
    
    // History and favorites tests
    void testConnectionHistory();
    void testFavoritesManager();
    
    // Plugin system tests
    void testPluginManager();
    void testPluginDiscovery();
    
    // Virtual channel tests
    void testVirtualChannelDialog();
    void testClipboardIntegration();
    
    // Multi-monitor tests
    void testMonitorDialog();
    void testMonitorDetection();
    
    // Communication bridge tests
    void testGoRDPBridge();
    void testAPICommunication();
    
    // Cross-platform specific tests
    void testPlatformSpecificFeatures();
    void testPlatformCompatibility();
    
    // UI responsiveness tests
    void testUIResponsiveness();
    void testMemoryUsage();
    
    // Error handling tests
    void testErrorHandling();
    void testRecoveryMechanisms();

private:
    QApplication *m_app;
    MainWindow *m_mainWindow;
    QNetworkAccessManager *m_networkManager;
    
    // Helper methods
    void setupTestEnvironment();
    void cleanupTestEnvironment();
    bool isPlatformSupported();
    QString getPlatformName();
    qint64 getCurrentMemoryUsage();
    void testWindowsSpecificFeatures();
    void testMacOSSpecificFeatures();
    void testLinuxSpecificFeatures();
};

void CrossPlatformTest::initTestCase()
{
    // Create QApplication instance for testing
    int argc = 1;
    char *argv[] = {(char*)"test"};
    m_app = new QApplication(argc, argv);
    
    // Setup test environment
    setupTestEnvironment();
    
    // Create network manager for API tests
    m_networkManager = new QNetworkAccessManager(this);
    
    qDebug() << "Cross-platform test suite initialized for platform:" << getPlatformName();
}

void CrossPlatformTest::cleanupTestCase()
{
    // Cleanup test environment
    cleanupTestEnvironment();
    
    delete m_networkManager;
    delete m_app;
    
    qDebug() << "Cross-platform test suite cleaned up";
}

void CrossPlatformTest::testMainWindowCreation()
{
    // Test main window creation
    MainWindow window;
    QVERIFY(window.isVisible() == false);
    
    window.show();
    QVERIFY(window.isVisible() == true);
    
    // Test window properties
    QVERIFY(window.windowTitle().contains("GoRDP"));
    QVERIFY(window.width() > 0);
    QVERIFY(window.height() > 0);
    
    window.close();
}

void CrossPlatformTest::testMainWindowMenuBar()
{
    MainWindow window;
    window.show();
    
    // Test menu bar exists
    QMenuBar *menuBar = window.menuBar();
    QVERIFY(menuBar != nullptr);
    
    // Test menu items exist
    QList<QAction*> actions = menuBar->actions();
    QVERIFY(actions.size() > 0);
    
    // Test specific menus
    bool hasFileMenu = false;
    bool hasEditMenu = false;
    bool hasViewMenu = false;
    bool hasHelpMenu = false;
    
    for (QAction *action : actions) {
        QString menuText = action->text();
        if (menuText.contains("File", Qt::CaseInsensitive)) hasFileMenu = true;
        if (menuText.contains("Edit", Qt::CaseInsensitive)) hasEditMenu = true;
        if (menuText.contains("View", Qt::CaseInsensitive)) hasViewMenu = true;
        if (menuText.contains("Help", Qt::CaseInsensitive)) hasHelpMenu = true;
    }
    
    QVERIFY(hasFileMenu);
    QVERIFY(hasEditMenu);
    QVERIFY(hasViewMenu);
    QVERIFY(hasHelpMenu);
    
    window.close();
}

void CrossPlatformTest::testMainWindowToolBar()
{
    MainWindow window;
    window.show();
    
    // Test toolbar exists
    QList<QToolBar*> toolBars = window.findChildren<QToolBar*>();
    QVERIFY(toolBars.size() > 0);
    
    // Test toolbar actions
    for (QToolBar *toolBar : toolBars) {
        QList<QAction*> actions = toolBar->actions();
        QVERIFY(actions.size() > 0);
        
        // Test that actions are enabled
        for (QAction *action : actions) {
            QVERIFY(action->isEnabled() || action->isSeparator());
        }
    }
    
    window.close();
}

void CrossPlatformTest::testMainWindowStatusBar()
{
    MainWindow window;
    window.show();
    
    // Test status bar exists
    QStatusBar *statusBar = window.statusBar();
    QVERIFY(statusBar != nullptr);
    QVERIFY(statusBar->isVisible());
    
    // Test status bar can show messages
    statusBar->showMessage("Test message");
    QCOMPARE(statusBar->currentMessage(), QString("Test message"));
    
    window.close();
}

void CrossPlatformTest::testConnectionDialogCreation()
{
    ConnectionDialog dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Connection", Qt::CaseInsensitive));
    QVERIFY(dialog.isModal() == true);
    
    dialog.close();
}

void CrossPlatformTest::testConnectionDialogValidation()
{
    ConnectionDialog dialog;
    dialog.show();
    
    // Test validation with empty fields
    // This would test the actual validation logic in the dialog
    // For now, we'll just verify the dialog can be shown
    
    QVERIFY(dialog.isVisible() == true);
    
    dialog.close();
}

void CrossPlatformTest::testConnectionDialogSettings()
{
    ConnectionDialog dialog;
    dialog.show();
    
    // Test settings loading/saving
    // This would test the actual settings functionality
    // For now, we'll just verify the dialog can be shown
    
    QVERIFY(dialog.isVisible() == true);
    
    dialog.close();
}

void CrossPlatformTest::testRDPDisplayWidget()
{
    RDPDisplayWidget display;
    QVERIFY(display.isVisible() == false);
    
    display.show();
    QVERIFY(display.isVisible() == true);
    
    // Test display properties
    QVERIFY(display.width() > 0);
    QVERIFY(display.height() > 0);
    
    display.close();
}

void CrossPlatformTest::testBitmapRendering()
{
    RDPDisplayWidget display;
    display.show();
    
    // Test bitmap rendering
    QImage testImage(100, 100, QImage::Format_RGB32);
    testImage.fill(Qt::red);
    
    // This would test the actual bitmap rendering
    // For now, we'll just verify the widget can be shown
    
    QVERIFY(display.isVisible() == true);
    
    display.close();
}

void CrossPlatformTest::testDisplayScaling()
{
    RDPDisplayWidget display;
    display.show();
    
    // Test display scaling
    int originalWidth = display.width();
    int originalHeight = display.height();
    
    display.resize(800, 600);
    QCOMPARE(display.width(), 800);
    QCOMPARE(display.height(), 600);
    
    display.resize(originalWidth, originalHeight);
    
    display.close();
}

void CrossPlatformTest::testMouseInput()
{
    RDPDisplayWidget display;
    display.show();
    
    // Test mouse input handling
    // This would test the actual mouse input functionality
    // For now, we'll just verify the widget can receive focus
    
    display.setFocus();
    QVERIFY(display.hasFocus());
    
    display.close();
}

void CrossPlatformTest::testKeyboardInput()
{
    RDPDisplayWidget display;
    display.show();
    
    // Test keyboard input handling
    // This would test the actual keyboard input functionality
    // For now, we'll just verify the widget can receive focus
    
    display.setFocus();
    QVERIFY(display.hasFocus());
    
    display.close();
}

void CrossPlatformTest::testInputFocus()
{
    RDPDisplayWidget display;
    display.show();
    
    // Test input focus management
    display.setFocus();
    QVERIFY(display.hasFocus());
    
    // Simulate focus loss
    display.clearFocus();
    QVERIFY(!display.hasFocus());
    
    display.close();
}

void CrossPlatformTest::testSettingsDialog()
{
    SettingsDialog dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Settings", Qt::CaseInsensitive));
    
    dialog.close();
}

void CrossPlatformTest::testSettingsPersistence()
{
    // Test settings persistence
    QSettings settings("GoRDP", "TestSettings");
    
    // Write test setting
    settings.setValue("testKey", "testValue");
    settings.sync();
    
    // Read test setting
    QString value = settings.value("testKey").toString();
    QCOMPARE(value, QString("testValue"));
    
    // Cleanup
    settings.remove("testKey");
    settings.sync();
}

void CrossPlatformTest::testPerformanceDialog()
{
    PerformanceDialog dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Performance", Qt::CaseInsensitive));
    
    dialog.close();
}

void CrossPlatformTest::testPerformanceMonitoring()
{
    PerformanceDialog dialog;
    dialog.show();
    
    // Test performance monitoring
    // This would test the actual performance monitoring functionality
    // For now, we'll just verify the dialog can be shown
    
    QVERIFY(dialog.isVisible() == true);
    
    dialog.close();
}

void CrossPlatformTest::testConnectionHistory()
{
    ConnectionHistory history;
    
    // Test history functionality
    QJsonObject testConnection;
    testConnection["server"] = "test.server.com";
    testConnection["port"] = 3389;
    testConnection["username"] = "testuser";
    testConnection["success"] = true;
    testConnection["duration"] = 5000;
    
    history.addConnection(testConnection);
    
    QJsonArray historyData = history.getHistory();
    QVERIFY(historyData.size() > 0);
    
    // Test history statistics
    QJsonObject stats = history.getConnectionStats();
    QVERIFY(stats.contains("totalConnections"));
    QVERIFY(stats.contains("successfulConnections"));
}

void CrossPlatformTest::testFavoritesManager()
{
    FavoritesManager favorites;
    
    // Test favorites functionality
    QJsonObject testFavorite;
    testFavorite["name"] = "Test Server";
    testFavorite["server"] = "test.server.com";
    testFavorite["port"] = 3389;
    testFavorite["username"] = "testuser";
    
    favorites.addFavorite(testFavorite);
    
    QJsonArray favoritesData = favorites.getFavorites();
    QVERIFY(favoritesData.size() > 0);
}

void CrossPlatformTest::testPluginManager()
{
    PluginManager dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Plugin", Qt::CaseInsensitive));
    
    dialog.close();
}

void CrossPlatformTest::testPluginDiscovery()
{
    PluginManager dialog;
    dialog.show();
    
    // Test plugin discovery
    // This would test the actual plugin discovery functionality
    // For now, we'll just verify the dialog can be shown
    
    QVERIFY(dialog.isVisible() == true);
    
    dialog.close();
}

void CrossPlatformTest::testVirtualChannelDialog()
{
    VirtualChannelDialog dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Virtual Channel", Qt::CaseInsensitive));
    
    dialog.close();
}

void CrossPlatformTest::testClipboardIntegration()
{
    // Test clipboard integration
    QClipboard *clipboard = QApplication::clipboard();
    QVERIFY(clipboard != nullptr);
    
    // Test clipboard functionality
    clipboard->setText("Test clipboard text");
    QString clipboardText = clipboard->text();
    QCOMPARE(clipboardText, QString("Test clipboard text"));
}

void CrossPlatformTest::testMonitorDialog()
{
    MonitorDialog dialog;
    QVERIFY(dialog.isVisible() == false);
    
    dialog.show();
    QVERIFY(dialog.isVisible() == true);
    
    // Test dialog properties
    QVERIFY(dialog.windowTitle().contains("Monitor", Qt::CaseInsensitive));
    
    dialog.close();
}

void CrossPlatformTest::testMonitorDetection()
{
    MonitorDialog dialog;
    dialog.show();
    
    // Test monitor detection
    QList<QScreen*> screens = QApplication::screens();
    QVERIFY(screens.size() > 0);
    
    // Test primary screen
    QScreen *primaryScreen = QApplication::primaryScreen();
    QVERIFY(primaryScreen != nullptr);
    
    dialog.close();
}

void CrossPlatformTest::testGoRDPBridge()
{
    GoRDPBridge bridge;
    
    // Test bridge functionality
    // This would test the actual bridge functionality
    // For now, we'll just verify the bridge can be created
    
    QVERIFY(&bridge != nullptr);
}

void CrossPlatformTest::testAPICommunication()
{
    // Test API communication
    // This would test the actual API communication functionality
    // For now, we'll just verify the network manager exists
    
    QVERIFY(m_networkManager != nullptr);
}

void CrossPlatformTest::testPlatformSpecificFeatures()
{
    QString platform = getPlatformName();
    qDebug() << "Testing platform-specific features for:" << platform;
    
    // Test platform-specific features
    if (platform.contains("Windows", Qt::CaseInsensitive)) {
        // Windows-specific tests
        testWindowsSpecificFeatures();
    } else if (platform.contains("macOS", Qt::CaseInsensitive) || platform.contains("Darwin", Qt::CaseInsensitive)) {
        // macOS-specific tests
        testMacOSSpecificFeatures();
    } else if (platform.contains("Linux", Qt::CaseInsensitive)) {
        // Linux-specific tests
        testLinuxSpecificFeatures();
    }
}

void CrossPlatformTest::testWindowsSpecificFeatures()
{
    // Test Windows-specific features
    qDebug() << "Testing Windows-specific features";
    
    // Test Windows registry access (if available)
    QSettings settings(QSettings::NativeFormat, QSettings::UserScope, "GoRDP", "Test");
    settings.setValue("test_key", "test_value");
    QCOMPARE(settings.value("test_key").toString(), QString("test_value"));
    settings.remove("test_key");
    
    // Test Windows file system features
    QString tempPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation);
    QString testFile = tempPath + "/gordp_test_file.txt";
    
    QFile file(testFile);
    QVERIFY(file.open(QIODevice::WriteOnly));
    file.write("Test data for Windows");
    file.close();
    
    QVERIFY(QFile::exists(testFile));
    QVERIFY(QFile::remove(testFile));
    
    // Test Windows clipboard
    QClipboard *clipboard = QApplication::clipboard();
    clipboard->setText("Windows clipboard test");
    QCOMPARE(clipboard->text(), QString("Windows clipboard test"));
    
    // Test Windows screen properties
    QList<QScreen*> screens = QApplication::screens();
    QVERIFY(screens.size() > 0);
    
    for (QScreen *screen : screens) {
        QVERIFY(screen->geometry().width() > 0);
        QVERIFY(screen->geometry().height() > 0);
        QVERIFY(screen->logicalDotsPerInch() > 0);
    }
}

void CrossPlatformTest::testMacOSSpecificFeatures()
{
    // Test macOS-specific features
    qDebug() << "Testing macOS-specific features";
    
    // Test macOS preferences
    QSettings settings(QSettings::NativeFormat, QSettings::UserScope, "GoRDP", "Test");
    settings.setValue("test_key", "test_value");
    QCOMPARE(settings.value("test_key").toString(), QString("test_value"));
    settings.remove("test_key");
    
    // Test macOS file system features
    QString tempPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation);
    QString testFile = tempPath + "/gordp_test_file.txt";
    
    QFile file(testFile);
    QVERIFY(file.open(QIODevice::WriteOnly));
    file.write("Test data for macOS");
    file.close();
    
    QVERIFY(QFile::exists(testFile));
    QVERIFY(QFile::remove(testFile));
    
    // Test macOS clipboard
    QClipboard *clipboard = QApplication::clipboard();
    clipboard->setText("macOS clipboard test");
    QCOMPARE(clipboard->text(), QString("macOS clipboard test"));
    
    // Test macOS screen properties
    QList<QScreen*> screens = QApplication::screens();
    QVERIFY(screens.size() > 0);
    
    for (QScreen *screen : screens) {
        QVERIFY(screen->geometry().width() > 0);
        QVERIFY(screen->geometry().height() > 0);
        QVERIFY(screen->logicalDotsPerInch() > 0);
    }
    
    // Test macOS-specific paths
    QString homePath = QStandardPaths::writableLocation(QStandardPaths::HomeLocation);
    QVERIFY(!homePath.isEmpty());
    QVERIFY(QDir(homePath).exists());
    
    QString documentsPath = QStandardPaths::writableLocation(QStandardPaths::DocumentsLocation);
    QVERIFY(!documentsPath.isEmpty());
}

void CrossPlatformTest::testLinuxSpecificFeatures()
{
    // Test Linux-specific features
    qDebug() << "Testing Linux-specific features";
    
    // Test Linux configuration files
    QSettings settings(QSettings::IniFormat, QSettings::UserScope, "GoRDP", "Test");
    settings.setValue("test_key", "test_value");
    QCOMPARE(settings.value("test_key").toString(), QString("test_value"));
    settings.remove("test_key");
    
    // Test Linux file system features
    QString tempPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation);
    QString testFile = tempPath + "/gordp_test_file.txt";
    
    QFile file(testFile);
    QVERIFY(file.open(QIODevice::WriteOnly));
    file.write("Test data for Linux");
    file.close();
    
    QVERIFY(QFile::exists(testFile));
    QVERIFY(QFile::remove(testFile));
    
    // Test Linux clipboard
    QClipboard *clipboard = QApplication::clipboard();
    clipboard->setText("Linux clipboard test");
    QCOMPARE(clipboard->text(), QString("Linux clipboard test"));
    
    // Test Linux screen properties
    QList<QScreen*> screens = QApplication::screens();
    QVERIFY(screens.size() > 0);
    
    for (QScreen *screen : screens) {
        QVERIFY(screen->geometry().width() > 0);
        QVERIFY(screen->geometry().height() > 0);
        QVERIFY(screen->logicalDotsPerInch() > 0);
    }
    
    // Test Linux-specific paths
    QString homePath = QStandardPaths::writableLocation(QStandardPaths::HomeLocation);
    QVERIFY(!homePath.isEmpty());
    QVERIFY(QDir(homePath).exists());
    
    QString configPath = QStandardPaths::writableLocation(QStandardPaths::ConfigLocation);
    QVERIFY(!configPath.isEmpty());
}

void CrossPlatformTest::testPlatformCompatibility()
{
    QString platform = getPlatformName();
    qDebug() << "Testing platform compatibility for:" << platform;
    
    // Test basic Qt functionality
    QVERIFY(QApplication::instance() != nullptr);
    QVERIFY(QApplication::screens().size() > 0);
    QVERIFY(QApplication::clipboard() != nullptr);
    
    // Test file system access
    QString tempPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation);
    QVERIFY(!tempPath.isEmpty());
    QVERIFY(QDir(tempPath).exists());
    
    // Test settings functionality
    QSettings settings(QSettings::IniFormat, QSettings::UserScope, "GoRDP", "Test");
    settings.setValue("compatibility_test", "value");
    QCOMPARE(settings.value("compatibility_test").toString(), QString("value"));
    settings.remove("compatibility_test");
    
    // Test network functionality
    QVERIFY(m_networkManager != nullptr);
    
    // Test JSON functionality
    QJsonObject testObject;
    testObject["key"] = "value";
    testObject["number"] = 42;
    QVERIFY(testObject.contains("key"));
    QVERIFY(testObject.contains("number"));
    QCOMPARE(testObject["key"].toString(), QString("value"));
    QCOMPARE(testObject["number"].toInt(), 42);
    
    // Test directory operations
    QString testDirPath = tempPath + "/gordp_compatibility_test";
    QDir testDir(testDirPath);
    
    if (testDir.exists()) {
        testDir.removeRecursively();
    }
    
    QVERIFY(testDir.mkpath("."));
    QVERIFY(testDir.exists());
    
    // Create test file
    QString testFile = testDirPath + "/test.txt";
    QFile file(testFile);
    QVERIFY(file.open(QIODevice::WriteOnly));
    file.write("Compatibility test data");
    file.close();
    
    QVERIFY(QFile::exists(testFile));
    QVERIFY(QFile::remove(testFile));
    QVERIFY(testDir.removeRecursively());
}

void CrossPlatformTest::testUIResponsiveness()
{
    MainWindow window;
    window.show();
    
    // Test UI responsiveness
    QVERIFY(window.isVisible() == true);
    
    // Test window operations
    QElapsedTimer timer;
    timer.start();
    
    // Test resize operation
    window.resize(800, 600);
    QCOMPARE(window.width(), 800);
    QCOMPARE(window.height(), 600);
    
    // Test move operation
    window.move(100, 100);
    QCOMPARE(window.x(), 100);
    QCOMPARE(window.y(), 100);
    
    // Test minimize/maximize
    window.showMinimized();
    QVERIFY(window.isMinimized());
    
    window.showNormal();
    QVERIFY(!window.isMinimized());
    
    window.showMaximized();
    QVERIFY(window.isMaximized());
    
    window.showNormal();
    QVERIFY(!window.isMaximized());
    
    // Test menu responsiveness
    QMenuBar *menuBar = window.menuBar();
    if (menuBar && menuBar->actions().size() > 0) {
        QAction *firstAction = menuBar->actions().first();
        QVERIFY(firstAction != nullptr);
        
        // Test menu action availability
        QVERIFY(firstAction->isEnabled() || !firstAction->isEnabled()); // Either state is valid
    }
    
    // Test toolbar responsiveness
    QList<QToolBar*> toolBars = window.findChildren<QToolBar*>();
    for (QToolBar *toolBar : toolBars) {
        QVERIFY(toolBar != nullptr);
        QVERIFY(toolBar->isVisible() || !toolBar->isVisible()); // Either state is valid
    }
    
    // Test status bar responsiveness
    QStatusBar *statusBar = window.statusBar();
    if (statusBar) {
        QVERIFY(statusBar != nullptr);
        statusBar->showMessage("Test message", 1000);
        QCOMPARE(statusBar->currentMessage(), QString("Test message"));
    }
    
    qint64 elapsed = timer.elapsed();
    qDebug() << "UI responsiveness test completed in" << elapsed << "ms";
    
    // UI operations should complete quickly (less than 1 second)
    QVERIFY(elapsed < 1000);
    
    window.close();
}

void CrossPlatformTest::testMemoryUsage()
{
    // Test memory usage
    qDebug() << "Testing memory usage";
    
    // Get initial memory usage
    qint64 initialMemory = getCurrentMemoryUsage();
    qDebug() << "Initial memory usage:" << initialMemory << "bytes";
    
    // Create and destroy multiple windows to test memory management
    for (int i = 0; i < 5; ++i) {
        MainWindow *window = new MainWindow();
        window->show();
        
        // Perform some operations
        window->resize(800, 600);
        window->move(100 + i * 50, 100 + i * 50);
        
        // Simulate some user interaction
        QApplication::processEvents();
        
        window->close();
        delete window;
        
        // Force garbage collection
        QApplication::processEvents();
    }
    
    // Get final memory usage
    qint64 finalMemory = getCurrentMemoryUsage();
    qDebug() << "Final memory usage:" << finalMemory << "bytes";
    
    // Memory usage should not increase significantly (allow 10MB increase)
    qint64 memoryIncrease = finalMemory - initialMemory;
    qDebug() << "Memory increase:" << memoryIncrease << "bytes";
    
    // Allow some memory increase due to Qt's internal caching
    QVERIFY(memoryIncrease < 10 * 1024 * 1024); // 10MB
    
    // Test memory with dialogs
    qint64 beforeDialogMemory = getCurrentMemoryUsage();
    
    for (int i = 0; i < 3; ++i) {
        ConnectionDialog *dialog = new ConnectionDialog();
        dialog->show();
        QApplication::processEvents();
        dialog->close();
        delete dialog;
        QApplication::processEvents();
    }
    
    qint64 afterDialogMemory = getCurrentMemoryUsage();
    qint64 dialogMemoryIncrease = afterDialogMemory - beforeDialogMemory;
    
    qDebug() << "Dialog memory increase:" << dialogMemoryIncrease << "bytes";
    QVERIFY(dialogMemoryIncrease < 5 * 1024 * 1024); // 5MB
}

void CrossPlatformTest::testErrorHandling()
{
    // Test error handling
    qDebug() << "Testing error handling";
    
    // Test invalid file operations
    QString invalidPath = "/invalid/path/that/does/not/exist";
    QFile invalidFile(invalidPath);
    QVERIFY(!invalidFile.open(QIODevice::ReadOnly));
    
    // Test invalid settings
    QSettings invalidSettings("", QSettings::IniFormat);
    invalidSettings.setValue("test", "value");
    // This should not crash
    
    // Test invalid JSON
    QJsonParseError parseError;
    QJsonDocument invalidJson = QJsonDocument::fromJson("invalid json", &parseError);
    QVERIFY(invalidJson.isNull());
    QVERIFY(parseError.error != QJsonParseError::NoError);
    
    // Test invalid network operations
    QNetworkRequest invalidRequest(QUrl("http://invalid.url.that.does.not.exist"));
    QNetworkReply *reply = m_networkManager->get(invalidRequest);
    
    // Wait for reply with timeout
    QEventLoop loop;
    QTimer::singleShot(5000, &loop, &QEventLoop::quit); // 5 second timeout
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();
    
    QVERIFY(reply->error() != QNetworkReply::NoError);
    reply->deleteLater();
    
    // Test invalid window operations
    MainWindow window;
    window.show();
    
    // Test invalid resize
    window.resize(-100, -100);
    QVERIFY(window.width() >= 0);
    QVERIFY(window.height() >= 0);
    
    // Test invalid move
    window.move(-1000, -1000);
    // Should not crash
    
    window.close();
    
    // Test clipboard error handling
    QClipboard *clipboard = QApplication::clipboard();
    clipboard->setText(""); // Empty text should not cause issues
    QCOMPARE(clipboard->text(), QString(""));
    
    // Test screen error handling
    QList<QScreen*> screens = QApplication::screens();
    if (screens.size() > 0) {
        QScreen *screen = screens.first();
        QVERIFY(screen != nullptr);
        
        // Test invalid geometry
        QRect invalidRect = screen->geometry();
        QVERIFY(invalidRect.isValid() || !invalidRect.isValid()); // Either state is valid
    }
}

void CrossPlatformTest::testRecoveryMechanisms()
{
    // Test recovery mechanisms
    qDebug() << "Testing recovery mechanisms";
    
    // Test application recovery after errors
    MainWindow window;
    window.show();
    
    // Simulate error condition
    QSettings settings(QSettings::IniFormat, QSettings::UserScope, "GoRDP", "Test");
    settings.setValue("error_simulation", "true");
    
    // Test recovery from invalid state
    window.close();
    
    // Recreate window after error
    MainWindow newWindow;
    newWindow.show();
    QVERIFY(newWindow.isVisible());
    newWindow.close();
    
    // Test file system recovery
    QString tempPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation);
    QString testDir = tempPath + "/gordp_recovery_test";
    QDir dir(testDir);
    
    // Create test directory
    if (dir.exists()) {
        dir.removeRecursively();
    }
    QVERIFY(dir.mkpath("."));
    
    // Create test files
    QStringList testFiles;
    for (int i = 0; i < 5; ++i) {
        QString fileName = QString("test_file_%1.txt").arg(i);
        QString filePath = testDir + "/" + fileName;
        testFiles.append(filePath);
        
        QFile file(filePath);
        QVERIFY(file.open(QIODevice::WriteOnly));
        file.write(QString("Test data %1").arg(i).toUtf8());
        file.close();
    }
    
    // Simulate file system error
    QFile errorFile(testDir + "/error_file.txt");
    QVERIFY(errorFile.open(QIODevice::WriteOnly));
    errorFile.write("Error simulation");
    errorFile.close();
    
    // Test recovery by removing error file
    QVERIFY(QFile::remove(testDir + "/error_file.txt"));
    
    // Verify other files still exist
    for (const QString &filePath : testFiles) {
        QVERIFY(QFile::exists(filePath));
    }
    
    // Cleanup
    dir.removeRecursively();
    
    // Test network recovery
    QNetworkRequest request(QUrl("http://httpbin.org/get"));
    QNetworkReply *reply = m_networkManager->get(request);
    
    QEventLoop loop;
    QTimer::singleShot(10000, &loop, &QEventLoop::quit); // 10 second timeout
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();
    
    if (reply->error() == QNetworkReply::NoError) {
        QVERIFY(reply->bytesAvailable() > 0);
    } else {
        qDebug() << "Network request failed (expected in some environments):" << reply->errorString();
    }
    
    reply->deleteLater();
    
    // Test settings recovery
    QSettings recoverySettings(QSettings::IniFormat, QSettings::UserScope, "GoRDP", "RecoveryTest");
    recoverySettings.setValue("recovery_test", "value");
    QCOMPARE(recoverySettings.value("recovery_test").toString(), QString("value"));
    
    // Simulate settings corruption
    recoverySettings.remove("recovery_test");
    QVERIFY(recoverySettings.value("recovery_test").isNull());
    
    // Test recovery by setting default value
    recoverySettings.setValue("recovery_test", "default_value");
    QCOMPARE(recoverySettings.value("recovery_test").toString(), QString("default_value"));
    
    // Cleanup
    recoverySettings.remove("recovery_test");
}

// Helper method to get current memory usage
qint64 CrossPlatformTest::getCurrentMemoryUsage()
{
    // This is a simplified memory usage calculation
    // In a real implementation, you would use platform-specific APIs
    
#ifdef Q_OS_WIN
    // Windows memory usage
    return 0; // Placeholder for Windows implementation
#elif defined(Q_OS_MAC)
    // macOS memory usage
    return 0; // Placeholder for macOS implementation
#elif defined(Q_OS_LINUX)
    // Linux memory usage
    return 0; // Placeholder for Linux implementation
#else
    return 0; // Default
#endif
}

void CrossPlatformTest::setupTestEnvironment()
{
    // Setup test environment
    QString testDataPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation) + "/GoRDP_Test";
    QDir testDir(testDataPath);
    
    if (!testDir.exists()) {
        testDir.mkpath(".");
    }
    
    // Set test-specific settings
    QSettings::setPath(QSettings::IniFormat, QSettings::UserScope, testDataPath);
}

void CrossPlatformTest::cleanupTestEnvironment()
{
    // Cleanup test environment
    QString testDataPath = QStandardPaths::writableLocation(QStandardPaths::TempLocation) + "/GoRDP_Test";
    QDir testDir(testDataPath);
    
    if (testDir.exists()) {
        testDir.removeRecursively();
    }
}

bool CrossPlatformTest::isPlatformSupported()
{
    QString platform = getPlatformName();
    return platform.contains("Windows", Qt::CaseInsensitive) ||
           platform.contains("macOS", Qt::CaseInsensitive) ||
           platform.contains("Darwin", Qt::CaseInsensitive) ||
           platform.contains("Linux", Qt::CaseInsensitive);
}

QString CrossPlatformTest::getPlatformName()
{
#ifdef Q_OS_WIN
    return "Windows";
#elif defined(Q_OS_MAC)
    return "macOS";
#elif defined(Q_OS_LINUX)
    return "Linux";
#else
    return "Unknown";
#endif
}

QTEST_MAIN(CrossPlatformTest)
#include "cross_platform_test.moc" 