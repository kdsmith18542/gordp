#include "mainwindow.h"
#include <QApplication>
#include <QMenuBar>
#include <QToolBar>
#include <QStatusBar>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QWidget>
#include <QAction>
#include <QMenu>
#include <QMessageBox>
#include <QSettings>
#include <QCloseEvent>
#include <QTimer>
#include <QIcon>
#include <QStyle>
#include <QDebug>
#include <QScreen>
#include <QActionGroup>
#include <QInputDialog>
#include <QApplication>
#include "../history/connection_history.h"
#include "../favorites/favorites_manager.h"

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_displayWidget(nullptr)
    , m_gordpBridge(new GoRDPBridge(this))
    , m_connectionDialog(nullptr)
    , m_settingsDialog(nullptr)
    , m_performanceDialog(nullptr)
    , m_historyDialog(nullptr)
    , m_favoritesDialog(nullptr)
    , m_pluginManager(nullptr)
    , m_virtualChannelDialog(nullptr)
    , m_monitorDialog(nullptr)
    , m_connectAction(nullptr)
    , m_disconnectAction(nullptr)
    , m_settingsAction(nullptr)
    , m_performanceAction(nullptr)
    , m_historyAction(nullptr)
    , m_favoritesAction(nullptr)
    , m_pluginsAction(nullptr)
    , m_virtualChannelsAction(nullptr)
    , m_multiMonitorAction(nullptr)
    , m_aboutAction(nullptr)
    , m_fullscreenAction(nullptr)
    , m_qualityMenu(nullptr)
    , m_resolutionMenu(nullptr)
    , m_qualityActionGroup(nullptr)
    , m_resolutionActionGroup(nullptr)
    , m_isConnected(false)
    , m_isFullscreen(false)
    , m_currentQuality("High")
    , m_currentResolution("1024x768")
    , m_statusTimer(new QTimer(this))
    , m_recentConnectionsMenu(nullptr)
{
    setWindowTitle("GoRDP GUI - Remote Desktop Client");
    setMinimumSize(800, 600);
    resize(1024, 768);
    
    setupUI();
    createActions();
    setupMenuBar();
    setupToolBar();
    setupStatusBar();
    setupCentralWidget();
    
    // Connect GoRDP bridge signals
    connect(m_gordpBridge, &GoRDPBridge::connectionStatusChanged,
            this, &MainWindow::onConnectionStatusChanged);
    connect(m_gordpBridge, &GoRDPBridge::bitmapReceived,
            this, &MainWindow::onBitmapReceived);
    connect(m_gordpBridge, &GoRDPBridge::connectionError,
            this, &MainWindow::onErrorOccurred);
    connect(m_gordpBridge, &GoRDPBridge::apiStarted,
            this, [this]() { statusBar()->showMessage("GoRDP API started", 3000); });
    connect(m_gordpBridge, &GoRDPBridge::apiStopped,
            this, [this]() { statusBar()->showMessage("GoRDP API stopped", 3000); });
    connect(m_gordpBridge, &GoRDPBridge::apiError,
            this, &MainWindow::onErrorOccurred);
    
    // Setup status timer
    connect(m_statusTimer, &QTimer::timeout, this, [this]() {
        if (m_isConnected) {
            m_gordpBridge->getConnectionStatus();
        }
    });
    m_statusTimer->setInterval(5000); // Check every 5 seconds
    
    // Load settings
    loadSettings();
    
    // Start GoRDP API
    m_gordpBridge->startGoRDPAPI();
}

MainWindow::~MainWindow()
{
    saveSettings();
    m_gordpBridge->stopGoRDPAPI();
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    if (m_isConnected) {
        QMessageBox::StandardButton reply = QMessageBox::question(
            this, "Disconnect", 
            "You are currently connected to a remote server. Do you want to disconnect and close the application?",
            QMessageBox::Yes | QMessageBox::No
        );
        
        if (reply == QMessageBox::Yes) {
            m_gordpBridge->disconnectFromServer();
            event->accept();
        } else {
            event->ignore();
        }
    } else {
        event->accept();
    }
}

void MainWindow::setupUI()
{
    // Set window icon (if available)
    // setWindowIcon(QIcon(":/icons/app_icon.png"));
    
    // Set window properties
    setWindowState(Qt::WindowMaximized);
}

void MainWindow::createActions()
{
    // Connect action
    m_connectAction = new QAction("Connect", this);
    m_connectAction->setShortcut(QKeySequence::Open);
    m_connectAction->setStatusTip("Connect to a remote server");
    connect(m_connectAction, &QAction::triggered, this, &MainWindow::onConnectClicked);
    
    // Disconnect action
    m_disconnectAction = new QAction("Disconnect", this);
    m_disconnectAction->setShortcut(QKeySequence::Close);
    m_disconnectAction->setStatusTip("Disconnect from current server");
    m_disconnectAction->setEnabled(false);
    connect(m_disconnectAction, &QAction::triggered, this, &MainWindow::onDisconnectClicked);
    
    // Settings action
    m_settingsAction = new QAction("Settings", this);
    m_settingsAction->setShortcut(QKeySequence::Preferences);
    m_settingsAction->setStatusTip("Open application settings");
    connect(m_settingsAction, &QAction::triggered, this, &MainWindow::onSettingsClicked);
    
    // Performance action
    m_performanceAction = new QAction("Performance", this);
    m_performanceAction->setStatusTip("View performance statistics");
    connect(m_performanceAction, &QAction::triggered, this, &MainWindow::onPerformanceClicked);
    
    // History action
    m_historyAction = new QAction("Connection History", this);
    m_historyAction->setStatusTip("View connection history");
    connect(m_historyAction, &QAction::triggered, this, &MainWindow::onHistoryClicked);
    
    // Favorites action
    m_favoritesAction = new QAction("Favorites", this);
    m_favoritesAction->setStatusTip("Manage favorite servers");
    connect(m_favoritesAction, &QAction::triggered, this, &MainWindow::onFavoritesClicked);
    
    // Plugins action
    m_pluginsAction = new QAction("Plugins", this);
    m_pluginsAction->setStatusTip("Manage plugins");
    connect(m_pluginsAction, &QAction::triggered, this, &MainWindow::onPluginsClicked);
    
    // Virtual Channels action
    m_virtualChannelsAction = new QAction("Virtual Channels", this);
    m_virtualChannelsAction->setStatusTip("Configure virtual channels");
    connect(m_virtualChannelsAction, &QAction::triggered, this, &MainWindow::onVirtualChannelsClicked);
    
    // Multi-Monitor action
    m_multiMonitorAction = new QAction("Multi-Monitor", this);
    m_multiMonitorAction->setStatusTip("Configure multi-monitor settings");
    connect(m_multiMonitorAction, &QAction::triggered, this, &MainWindow::onMultiMonitorClicked);
    
    // Fullscreen action
    m_fullscreenAction = new QAction("Fullscreen", this);
    m_fullscreenAction->setShortcut(QKeySequence::FullScreen);
    m_fullscreenAction->setStatusTip("Toggle fullscreen mode");
    m_fullscreenAction->setCheckable(true);
    connect(m_fullscreenAction, &QAction::triggered, this, &MainWindow::onFullscreenToggled);
    
    // About action
    m_aboutAction = new QAction("About", this);
    m_aboutAction->setStatusTip("About GoRDP GUI");
    connect(m_aboutAction, &QAction::triggered, this, &MainWindow::onAboutClicked);
    
    // Create quality and resolution action groups
    m_qualityActionGroup = new QActionGroup(this);
    m_resolutionActionGroup = new QActionGroup(this);
    m_qualityActionGroup->setExclusive(true);
    m_resolutionActionGroup->setExclusive(true);
}

void MainWindow::setupMenuBar()
{
    QMenuBar *menuBar = this->menuBar();
    
    // File menu
    QMenu *fileMenu = menuBar->addMenu("&File");
    fileMenu->addAction(m_connectAction);
    fileMenu->addAction(m_disconnectAction);
    fileMenu->addSeparator();
    fileMenu->addAction(m_settingsAction);
    fileMenu->addSeparator();
    fileMenu->addAction("E&xit", this, &QWidget::close, QKeySequence::Quit);
    
    // Connection menu
    QMenu *connectionMenu = menuBar->addMenu("&Connection");
    connectionMenu->addAction(m_historyAction);
    connectionMenu->addAction(m_favoritesAction);
    connectionMenu->addSeparator();
    connectionMenu->addAction(m_performanceAction);
    connectionMenu->addAction(m_virtualChannelsAction);
    connectionMenu->addAction(m_multiMonitorAction);
    
    // View menu
    QMenu *viewMenu = menuBar->addMenu("&View");
    viewMenu->addAction(m_fullscreenAction);
    viewMenu->addSeparator();
    
    // Quality submenu
    m_qualityMenu = viewMenu->addMenu("&Quality");
    setupQualityMenu();
    
    // Resolution submenu
    m_resolutionMenu = viewMenu->addMenu("&Resolution");
    setupResolutionMenu();
    
    // Tools menu
    QMenu *toolsMenu = menuBar->addMenu("&Tools");
    toolsMenu->addAction(m_pluginsAction);
    
    // Help menu
    QMenu *helpMenu = menuBar->addMenu("&Help");
    helpMenu->addAction(m_aboutAction);
}

void MainWindow::setupQualityMenu()
{
    // Clear existing actions
    m_qualityMenu->clear();
    
    // Quality options
    QStringList qualities = {"Low", "Medium", "High", "Ultra"};
    
    for (const QString &quality : qualities) {
        QAction *action = new QAction(quality, this);
        action->setCheckable(true);
        action->setData(quality);
        m_qualityActionGroup->addAction(action);
        m_qualityMenu->addAction(action);
        
        if (quality == m_currentQuality) {
            action->setChecked(true);
        }
        
        connect(action, &QAction::triggered, this, [this, quality]() {
            setQuality(quality);
        });
    }
}

void MainWindow::setupResolutionMenu()
{
    // Clear existing actions
    m_resolutionMenu->clear();
    
    // Common resolutions
    QStringList resolutions = {
        "800x600", "1024x768", "1280x720", "1280x800", "1280x1024",
        "1366x768", "1440x900", "1600x900", "1680x1050", "1920x1080",
        "1920x1200", "2560x1440", "3840x2160"
    };
    
    for (const QString &resolution : resolutions) {
        QAction *action = new QAction(resolution, this);
        action->setCheckable(true);
        action->setData(resolution);
        m_resolutionActionGroup->addAction(action);
        m_resolutionMenu->addAction(action);
        
        if (resolution == m_currentResolution) {
            action->setChecked(true);
        }
        
        connect(action, &QAction::triggered, this, [this, resolution]() {
            setResolution(resolution);
        });
    }
    
    // Add custom resolution option
    m_resolutionMenu->addSeparator();
    QAction *customAction = new QAction("Custom...", this);
    connect(customAction, &QAction::triggered, this, &MainWindow::onCustomResolutionClicked);
    m_resolutionMenu->addAction(customAction);
}

void MainWindow::setupToolBar()
{
    QToolBar *toolBar = addToolBar("Main Toolbar");
    toolBar->setMovable(false);
    
    toolBar->addAction(m_connectAction);
    toolBar->addAction(m_disconnectAction);
    toolBar->addSeparator();
    toolBar->addAction(m_settingsAction);
    toolBar->addAction(m_performanceAction);
    toolBar->addSeparator();
    toolBar->addAction(m_fullscreenAction);
}

void MainWindow::setupStatusBar()
{
    statusBar()->showMessage("Ready");
}

void MainWindow::setupCentralWidget()
{
    QWidget *centralWidget = new QWidget(this);
    setCentralWidget(centralWidget);
    
    QVBoxLayout *mainLayout = new QVBoxLayout(centralWidget);
    mainLayout->setContentsMargins(0, 0, 0, 0);
    
    // Create RDP display widget
    m_displayWidget = new RDPDisplayWidget(this);
    mainLayout->addWidget(m_displayWidget);
    
    // Connect display widget signals to GoRDP bridge
    connect(m_displayWidget, &RDPDisplayWidget::mouseEvent,
            m_gordpBridge, &GoRDPBridge::sendMouseEvent);
    connect(m_displayWidget, &RDPDisplayWidget::keyEvent,
            m_gordpBridge, &GoRDPBridge::sendKeyEvent);
}

void MainWindow::updateRecentConnectionsMenu() {
    // Ensure the menu exists
    if (!m_recentConnectionsMenu) {
        m_recentConnectionsMenu = new QMenu("Recent Connections", this);
        menuBar()->addMenu(m_recentConnectionsMenu);
    }
    m_recentConnectionsMenu->clear();
    if (!m_connectionHistory) return;
    QJsonArray history = m_connectionHistory->getHistory();
    for (const QJsonValue& val : history) {
        QJsonObject conn = val.toObject();
        QString label = QString("%1@%2").arg(conn["username"].toString(), conn["server"].toString());
        QAction* action = new QAction(label, this);
        action->setData(conn);
        connect(action, &QAction::triggered, this, [this, conn]() {
            // Extract fields and connect
            QString server = conn["server"].toString();
            int port = conn["port"].toInt();
            QString username = conn["username"].toString();
            QString password = conn["password"].toString();
            QJsonObject options = conn.contains("options") ? conn["options"].toObject() : QJsonObject();
            m_gordpBridge->connectToServer(server, port, username, password, options);
        });
        m_recentConnectionsMenu->addAction(action);
    }
}

void MainWindow::loadSettings()
{
    QSettings settings;
    
    // Load window geometry
    restoreGeometry(settings.value("geometry").toByteArray());
    restoreState(settings.value("windowState").toByteArray());
    
    // Load quality and resolution settings
    m_currentQuality = settings.value("quality", "High").toString();
    m_currentResolution = settings.value("resolution", "1024x768").toString();
    
    // Update menus
    setupQualityMenu();
    setupResolutionMenu();
}

void MainWindow::saveSettings()
{
    QSettings settings;
    
    // Save window geometry
    settings.setValue("geometry", saveGeometry());
    settings.setValue("windowState", saveState());
    
    // Save quality and resolution settings
    settings.setValue("quality", m_currentQuality);
    settings.setValue("resolution", m_currentResolution);
}

void MainWindow::updateConnectionStatus(bool connected)
{
    m_isConnected = connected;
    
    // Update actions
    m_connectAction->setEnabled(!connected);
    m_disconnectAction->setEnabled(connected);
    
    // Update status bar
    if (connected) {
        statusBar()->showMessage("Connected to remote server");
        m_statusTimer->start();
    } else {
        statusBar()->showMessage("Disconnected");
        m_statusTimer->stop();
    }
}

void MainWindow::onConnectClicked()
{
    if (!m_connectionDialog) {
        m_connectionDialog = new ConnectionDialog(this);
        connect(m_connectionDialog, &ConnectionDialog::connectRequested,
                this, [this](const QString &server, int port, 
                            const QString &username, const QString &password,
                            const QJsonObject &options) {
            m_gordpBridge->connectToServer(server, port, username, password, options);
        });
    }
    
    m_connectionDialog->show();
    m_connectionDialog->raise();
    m_connectionDialog->activateWindow();
}

void MainWindow::onDisconnectClicked()
{
    m_gordpBridge->disconnectFromServer();
}

void MainWindow::onSettingsClicked()
{
    if (!m_settingsDialog) {
        m_settingsDialog = new SettingsDialog(this);
    }
    
    m_settingsDialog->show();
    m_settingsDialog->raise();
    m_settingsDialog->activateWindow();
}

void MainWindow::onPerformanceClicked()
{
    if (!m_performanceDialog) {
        m_performanceDialog = new PerformanceDialog(this);
    }
    
    m_performanceDialog->show();
    m_performanceDialog->raise();
    m_performanceDialog->activateWindow();
}

void MainWindow::onHistoryClicked()
{
    if (!m_historyDialog) {
        m_historyDialog = new HistoryDialog(this);
    }
    
    m_historyDialog->show();
    m_historyDialog->raise();
    m_historyDialog->activateWindow();
}

void MainWindow::onFavoritesClicked()
{
    if (!m_favoritesDialog) {
        m_favoritesDialog = new FavoritesDialog(this);
    }
    
    m_favoritesDialog->show();
    m_favoritesDialog->raise();
    m_favoritesDialog->activateWindow();
}

void MainWindow::onPluginsClicked()
{
    if (!m_pluginManager) {
        m_pluginManager = new PluginManager(this);
    }
    
    m_pluginManager->show();
    m_pluginManager->raise();
    m_pluginManager->activateWindow();
}

void MainWindow::onVirtualChannelsClicked()
{
    if (!m_virtualChannelDialog) {
        m_virtualChannelDialog = new VirtualChannelDialog(this);
    }
    
    m_virtualChannelDialog->show();
    m_virtualChannelDialog->raise();
    m_virtualChannelDialog->activateWindow();
}

void MainWindow::onMultiMonitorClicked()
{
    if (!m_monitorDialog) {
        m_monitorDialog = new MonitorDialog(this);
    }
    
    m_monitorDialog->show();
    m_monitorDialog->raise();
    m_monitorDialog->activateWindow();
}

void MainWindow::onAboutClicked()
{
    QMessageBox::about(this, "About GoRDP GUI",
        "<h3>GoRDP GUI</h3>"
        "<p>Version 1.0.0</p>"
        "<p>A modern Qt C++ GUI for the GoRDP remote desktop client.</p>"
        "<p>Built with Qt6 and C++17 for high performance and cross-platform compatibility.</p>"
        "<p>For more information, visit: <a href='https://github.com/gordp/gordp'>https://github.com/gordp/gordp</a></p>");
}

void MainWindow::onConnectionStatusChanged(bool connected)
{
    updateConnectionStatus(connected);
}

void MainWindow::onBitmapReceived(const QImage &image)
{
    if (m_displayWidget) {
        m_displayWidget->updateBitmap(image);
    }
}

void MainWindow::onErrorOccurred(const QString &error)
{
    QMessageBox::warning(this, "Error", error);
    statusBar()->showMessage("Error: " + error, 5000);
} 

void MainWindow::onFullscreenToggled()
{
    if (m_isFullscreen) {
        showNormal();
        m_isFullscreen = false;
        m_fullscreenAction->setChecked(false);
        statusBar()->showMessage("Exited fullscreen mode", 2000);
    } else {
        showFullScreen();
        m_isFullscreen = true;
        m_fullscreenAction->setChecked(true);
        statusBar()->showMessage("Entered fullscreen mode", 2000);
    }
}

void MainWindow::setQuality(const QString &quality)
{
    m_currentQuality = quality;
    
    // Update quality settings in GoRDP bridge
    if (m_gordpBridge) {
        QJsonObject qualitySettings;
        qualitySettings["quality"] = quality;
        m_gordpBridge->updateQualitySettings(qualitySettings);
    }
    
    // Update menu
    for (QAction *action : m_qualityActionGroup->actions()) {
        if (action->data().toString() == quality) {
            action->setChecked(true);
            break;
        }
    }
    
    statusBar()->showMessage(QString("Quality set to: %1").arg(quality), 2000);
    
    // Save setting
    QSettings settings;
    settings.setValue("quality", quality);
}

void MainWindow::setResolution(const QString &resolution)
{
    m_currentResolution = resolution;
    
    // Parse resolution
    QStringList parts = resolution.split('x');
    if (parts.size() == 2) {
        int width = parts[0].toInt();
        int height = parts[1].toInt();
        
        // Update display widget size
        if (m_displayWidget) {
            m_displayWidget->setFixedSize(width, height);
        }
        
        // Update GoRDP bridge resolution
        if (m_gordpBridge) {
            QJsonObject resolutionSettings;
            resolutionSettings["width"] = width;
            resolutionSettings["height"] = height;
            m_gordpBridge->updateResolutionSettings(resolutionSettings);
        }
    }
    
    // Update menu
    for (QAction *action : m_resolutionActionGroup->actions()) {
        if (action->data().toString() == resolution) {
            action->setChecked(true);
            break;
        }
    }
    
    statusBar()->showMessage(QString("Resolution set to: %1").arg(resolution), 2000);
    
    // Save setting
    QSettings settings;
    settings.setValue("resolution", resolution);
}

void MainWindow::onCustomResolutionClicked()
{
    bool ok;
    QString resolution = QInputDialog::getText(
        this, "Custom Resolution",
        "Enter resolution (e.g., 1920x1080):",
        QLineEdit::Normal,
        m_currentResolution,
        &ok
    );
    
    if (ok && !resolution.isEmpty()) {
        // Validate format
        QStringList parts = resolution.split('x');
        if (parts.size() == 2) {
            bool widthOk, heightOk;
            int width = parts[0].toInt(&widthOk);
            int height = parts[1].toInt(&heightOk);
            
            if (widthOk && heightOk && width > 0 && height > 0) {
                setResolution(resolution);
            } else {
                QMessageBox::warning(this, "Invalid Resolution",
                    "Please enter a valid resolution in the format WIDTHxHEIGHT (e.g., 1920x1080)");
            }
        } else {
            QMessageBox::warning(this, "Invalid Format",
                "Please enter resolution in the format WIDTHxHEIGHT (e.g., 1920x1080)");
        }
    }
} 