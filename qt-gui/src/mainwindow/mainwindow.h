#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
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
#include <QJsonArray>
#include <QVariantList>
#include <QActionGroup>

#include "../display/rdp_display.h"
#include "../connection/connection_dialog.h"
#include "../settings/settings_dialog.h"
#include "../performance/performance_dialog.h"
#include "../history/history_dialog.h"
#include "../favorites/favorites_dialog.h"
#include "../plugins/plugin_manager.h"
#include "../virtualchannels/virtual_channel_dialog.h"
#include "../multimonitor/monitor_dialog.h"
#include "../utils/gordp_bridge.h"

// Forward declarations
class ConnectionHistory;
class FavoritesManager;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

protected:
    void closeEvent(QCloseEvent *event) override;

private slots:
    void onConnectClicked();
    void onDisconnectClicked();
    void onSettingsClicked();
    void onPerformanceClicked();
    void onHistoryClicked();
    void onFavoritesClicked();
    void onPluginsClicked();
    void onVirtualChannelsClicked();
    void onMultiMonitorClicked();
    void onAboutClicked();
    void onConnectionStatusChanged(bool connected);
    void onBitmapReceived(const QImage &image);
    void onErrorOccurred(const QString &error);
    void onFullscreenToggled();
    void onCustomResolutionClicked();

private:
    void setupUI();
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupCentralWidget();
    void createActions();
    void setupQualityMenu();
    void setupResolutionMenu();
    void loadSettings();
    void saveSettings();
    void updateConnectionStatus(bool connected);
    void updateRecentConnectionsMenu();
    void setQuality(const QString &quality);
    void setResolution(const QString &resolution);

    // UI Components
    RDPDisplayWidget *m_displayWidget;
    GoRDPBridge *m_gordpBridge;
    QMenu* m_recentConnectionsMenu;
    
    // Dialogs
    ConnectionDialog *m_connectionDialog;
    SettingsDialog *m_settingsDialog;
    PerformanceDialog *m_performanceDialog;
    HistoryDialog *m_historyDialog;
    FavoritesDialog *m_favoritesDialog;
    PluginManager *m_pluginManager;
    VirtualChannelDialog *m_virtualChannelDialog;
    MonitorDialog *m_monitorDialog;
    
    // Managers
    ConnectionHistory *m_connectionHistory;
    FavoritesManager *m_favoritesManager;
    
    // Actions
    QAction *m_connectAction;
    QAction *m_disconnectAction;
    QAction *m_settingsAction;
    QAction *m_performanceAction;
    QAction *m_historyAction;
    QAction *m_favoritesAction;
    QAction *m_pluginsAction;
    QAction *m_virtualChannelsAction;
    QAction *m_multiMonitorAction;
    QAction *m_aboutAction;
    QAction *m_fullscreenAction;
    
    // Menus and Action Groups
    QMenu *m_qualityMenu;
    QMenu *m_resolutionMenu;
    QActionGroup *m_qualityActionGroup;
    QActionGroup *m_resolutionActionGroup;
    
    // Status
    bool m_isConnected;
    bool m_isFullscreen;
    QString m_currentQuality;
    QString m_currentResolution;
    QTimer *m_statusTimer;
    QVariantList m_recentConnections;
};

#endif // MAINWINDOW_H 