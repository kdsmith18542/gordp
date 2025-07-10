#ifndef PLUGIN_MANAGER_H
#define PLUGIN_MANAGER_H

#include <QDialog>
#include <QTableWidget>
#include <QJsonArray>
#include <QJsonObject>
#include <QSettings>
#include <QString>

namespace Ui {
class PluginManager;
}

class PluginManager : public QDialog
{
    Q_OBJECT

public:
    explicit PluginManager(QWidget *parent = nullptr);
    ~PluginManager();

    // Plugin management
    void loadPlugins(const QJsonArray &plugins);
    void enablePlugin(const QString &pluginName);
    void disablePlugin(const QString &pluginName);

signals:
    void pluginEnabled(const QString &pluginName);
    void pluginDisabled(const QString &pluginName);
    void pluginConfigured(const QString &pluginName);

private slots:
    void onEnableClicked();
    void onDisableClicked();
    void onConfigureClicked();
    void onInstallClicked();
    void onCloseClicked();

private:
    void setupConnections();
    void setupTable();
    void loadPlugins();
    void loadSettings();
    void savePluginSettings();
    void updatePluginTable();
    QJsonObject getPluginInfo(const QString &pluginPath);
    QJsonObject extractPluginMetadata(const QString &pluginPath);
    QJsonObject extractMetadataFromBinary(const QString &pluginPath);
    bool validatePluginFile(const QString &pluginPath);
    bool installPlugin(const QString &pluginPath);
    void updateUI();
    
    Ui::PluginManager *ui;
    QJsonArray m_plugins;
    QSettings *m_settings;
};

#endif // PLUGIN_MANAGER_H 