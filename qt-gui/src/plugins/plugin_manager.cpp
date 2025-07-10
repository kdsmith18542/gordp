#include "plugin_manager.h"
#include "ui_plugin_manager.h"
#include <QMessageBox>
#include <QJsonDocument>
#include <QDir>
#include <QFileDialog>
#include <QStandardPaths>
#include <QProcess>
#include <QDebug>
#include <QRegularExpression>

PluginManager::PluginManager(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::PluginManager)
    , m_settings(new QSettings("GoRDP", "Plugins"))
{
    ui->setupUi(this);
    setupConnections();
    setupTable();
    loadPlugins();
    loadSettings();
}

PluginManager::~PluginManager()
{
    delete ui;
    delete m_settings;
}

void PluginManager::loadPlugins(const QJsonArray &plugins)
{
    m_plugins = plugins;
    updatePluginTable();
}

void PluginManager::enablePlugin(const QString &pluginName)
{
    // Find and enable the plugin
    for (int i = 0; i < m_plugins.size(); ++i) {
        QJsonObject plugin = m_plugins[i].toObject();
        if (plugin["name"].toString() == pluginName) {
            plugin["enabled"] = true;
            m_plugins[i] = plugin;
            break;
        }
    }
    
    // Save settings
    savePluginSettings();
    updatePluginTable();
    
    emit pluginEnabled(pluginName);
}

void PluginManager::disablePlugin(const QString &pluginName)
{
    // Find and disable the plugin
    for (int i = 0; i < m_plugins.size(); ++i) {
        QJsonObject plugin = m_plugins[i].toObject();
        if (plugin["name"].toString() == pluginName) {
            plugin["enabled"] = false;
            m_plugins[i] = plugin;
            break;
        }
    }
    
    // Save settings
    savePluginSettings();
    updatePluginTable();
    
    emit pluginDisabled(pluginName);
}

void PluginManager::onEnableClicked()
{
    QList<QTableWidgetItem*> selectedItems = ui->pluginTable->selectedItems();
    if (selectedItems.isEmpty()) {
        QMessageBox::warning(this, "No Plugin Selected", 
                           "Please select a plugin to enable.");
        return;
    }
    
    int row = selectedItems.first()->row();
    QString pluginName = ui->pluginTable->item(row, 0)->text();
    
    enablePlugin(pluginName);
    QMessageBox::information(this, "Plugin Enabled", 
                           QString("Plugin '%1' has been enabled.").arg(pluginName));
}

void PluginManager::onDisableClicked()
{
    QList<QTableWidgetItem*> selectedItems = ui->pluginTable->selectedItems();
    if (selectedItems.isEmpty()) {
        QMessageBox::warning(this, "No Plugin Selected", 
                           "Please select a plugin to disable.");
        return;
    }
    
    int row = selectedItems.first()->row();
    QString pluginName = ui->pluginTable->item(row, 0)->text();
    
    disablePlugin(pluginName);
    QMessageBox::information(this, "Plugin Disabled", 
                           QString("Plugin '%1' has been disabled.").arg(pluginName));
}

void PluginManager::onConfigureClicked()
{
    QList<QTableWidgetItem*> selectedItems = ui->pluginTable->selectedItems();
    if (selectedItems.isEmpty()) {
        QMessageBox::warning(this, "No Plugin Selected", 
                           "Please select a plugin to configure.");
        return;
    }
    
    int row = selectedItems.first()->row();
    QString pluginName = ui->pluginTable->item(row, 0)->text();
    
    // Find plugin configuration
    QJsonObject pluginConfig;
    for (const QJsonValue &value : m_plugins) {
        QJsonObject plugin = value.toObject();
        if (plugin["name"].toString() == pluginName) {
            pluginConfig = plugin;
            break;
        }
    }
    
    if (!pluginConfig.isEmpty()) {
        emit pluginConfigured(pluginName);
        
        // Show configuration dialog (in a real implementation, this would be a custom dialog)
        QMessageBox::information(this, "Plugin Configuration", 
                               QString("Configuration dialog for plugin '%1' would open here.").arg(pluginName));
    }
}

void PluginManager::onInstallClicked()
{
    QString pluginPath = QFileDialog::getOpenFileName(this, 
        "Select Plugin File", 
        QStandardPaths::writableLocation(QStandardPaths::HomeLocation),
        "Plugin Files (*.so *.dll *.dylib);;All Files (*)");
    
    if (pluginPath.isEmpty()) {
        return;
    }
    
    // Validate plugin file
    if (!validatePluginFile(pluginPath)) {
        QMessageBox::critical(this, "Invalid Plugin", 
                            "The selected file is not a valid GoRDP plugin.");
        return;
    }
    
    // Install plugin
    if (installPlugin(pluginPath)) {
        QMessageBox::information(this, "Plugin Installed", 
                               "Plugin has been successfully installed.");
        loadPlugins(); // Reload plugin list
    } else {
        QMessageBox::critical(this, "Installation Failed", 
                            "Failed to install the plugin. Please check permissions and try again.");
    }
}

void PluginManager::onCloseClicked()
{
    accept();
}

void PluginManager::setupConnections()
{
    connect(ui->enableButton, &QPushButton::clicked, this, &PluginManager::onEnableClicked);
    connect(ui->disableButton, &QPushButton::clicked, this, &PluginManager::onDisableClicked);
    connect(ui->configureButton, &QPushButton::clicked, this, &PluginManager::onConfigureClicked);
    connect(ui->installButton, &QPushButton::clicked, this, &PluginManager::onInstallClicked);
    connect(ui->closeButton, &QPushButton::clicked, this, &PluginManager::onCloseClicked);
}

void PluginManager::setupTable()
{
    // Setup table headers
    ui->pluginTable->setColumnCount(4);
    ui->pluginTable->setHorizontalHeaderLabels({
        "Name", "Version", "Status", "Description"
    });
    
    // Setup table properties
    ui->pluginTable->setSelectionBehavior(QAbstractItemView::SelectRows);
    ui->pluginTable->setSelectionMode(QAbstractItemView::SingleSelection);
    ui->pluginTable->setAlternatingRowColors(true);
    ui->pluginTable->setSortingEnabled(true);
    
    // Resize columns
    ui->pluginTable->horizontalHeader()->setStretchLastSection(true);
    ui->pluginTable->resizeColumnsToContents();
}

void PluginManager::loadPlugins()
{
    m_plugins = QJsonArray();
    
    // Get plugins directory
    QString pluginsDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/plugins";
    QDir dir(pluginsDir);
    
    if (!dir.exists()) {
        dir.mkpath(".");
        return;
    }
    
    // Scan for plugin files
    QStringList filters;
    filters << "*.so" << "*.dll" << "*.dylib";
    QFileInfoList pluginFiles = dir.entryInfoList(filters, QDir::Files);
    
    for (const QFileInfo &fileInfo : pluginFiles) {
        QString pluginPath = fileInfo.absoluteFilePath();
        
        // Extract plugin metadata
        QJsonObject pluginInfo = getPluginInfo(pluginPath);
        if (!pluginInfo.isEmpty()) {
            m_plugins.append(pluginInfo);
        }
    }
    
    updatePluginTable();
}

void PluginManager::loadSettings()
{
    // Load enabled/disabled state for plugins
    for (int i = 0; i < m_plugins.size(); ++i) {
        QJsonObject plugin = m_plugins[i].toObject();
        QString pluginName = plugin["name"].toString();
        
        bool enabled = m_settings->value(QString("plugins/%1/enabled").arg(pluginName), false).toBool();
        plugin["enabled"] = enabled;
        m_plugins[i] = plugin;
    }
    
    updatePluginTable();
}

void PluginManager::savePluginSettings()
{
    // Save enabled/disabled state for plugins
    for (const QJsonValue &value : m_plugins) {
        QJsonObject plugin = value.toObject();
        QString pluginName = plugin["name"].toString();
        bool enabled = plugin["enabled"].toBool();
        
        m_settings->setValue(QString("plugins/%1/enabled").arg(pluginName), enabled);
    }
    
    m_settings->sync();
}

void PluginManager::updatePluginTable()
{
    ui->pluginTable->setRowCount(m_plugins.size());
    
    for (int i = 0; i < m_plugins.size(); ++i) {
        QJsonObject plugin = m_plugins[i].toObject();
        
        // Name
        QTableWidgetItem *nameItem = new QTableWidgetItem(plugin["name"].toString());
        nameItem->setFlags(nameItem->flags() & ~Qt::ItemIsEditable);
        ui->pluginTable->setItem(i, 0, nameItem);
        
        // Version
        QTableWidgetItem *versionItem = new QTableWidgetItem(plugin["version"].toString());
        versionItem->setFlags(versionItem->flags() & ~Qt::ItemIsEditable);
        ui->pluginTable->setItem(i, 1, versionItem);
        
        // Status
        QString status = plugin["enabled"].toBool() ? "Enabled" : "Disabled";
        QTableWidgetItem *statusItem = new QTableWidgetItem(status);
        statusItem->setFlags(statusItem->flags() & ~Qt::ItemIsEditable);
        if (plugin["enabled"].toBool()) {
            statusItem->setBackground(QColor(200, 255, 200)); // Light green
        } else {
            statusItem->setBackground(QColor(255, 200, 200)); // Light red
        }
        ui->pluginTable->setItem(i, 2, statusItem);
        
        // Description
        QTableWidgetItem *descItem = new QTableWidgetItem(plugin["description"].toString());
        descItem->setFlags(descItem->flags() & ~Qt::ItemIsEditable);
        ui->pluginTable->setItem(i, 3, descItem);
    }
    
    ui->pluginTable->resizeColumnsToContents();
}

QJsonObject PluginManager::getPluginInfo(const QString &pluginPath)
{
    QFileInfo fileInfo(pluginPath);
    QString fileName = fileInfo.baseName();
    
    QJsonObject pluginInfo;
    pluginInfo["name"] = fileName;
    pluginInfo["path"] = pluginPath;
    pluginInfo["enabled"] = false;
    
    // Try to extract real metadata from plugin file
    QJsonObject metadata = extractPluginMetadata(pluginPath);
    if (!metadata.isEmpty()) {
        pluginInfo["version"] = metadata.value("version").toString("1.0.0");
        pluginInfo["description"] = metadata.value("description").toString(QString("Plugin: %1").arg(fileName));
        pluginInfo["author"] = metadata.value("author").toString("Unknown");
        pluginInfo["license"] = metadata.value("license").toString("Unknown");
        pluginInfo["api_version"] = metadata.value("api_version").toString("1.0");
        pluginInfo["dependencies"] = metadata.value("dependencies").toArray();
    } else {
        // Fallback to basic info
        pluginInfo["version"] = "1.0.0";
        pluginInfo["description"] = QString("Plugin: %1").arg(fileName);
        pluginInfo["author"] = "Unknown";
        pluginInfo["license"] = "Unknown";
        pluginInfo["api_version"] = "1.0";
        pluginInfo["dependencies"] = QJsonArray();
    }
    
    return pluginInfo;
}

QJsonObject PluginManager::extractPluginMetadata(const QString &pluginPath)
{
    QJsonObject metadata;
    
    // Try to read metadata from JSON file with same name
    QFileInfo fileInfo(pluginPath);
    QString metadataPath = fileInfo.absolutePath() + "/" + fileInfo.baseName() + ".json";
    QFile metadataFile(metadataPath);
    
    if (metadataFile.open(QIODevice::ReadOnly)) {
        QJsonDocument doc = QJsonDocument::fromJson(metadataFile.readAll());
        if (doc.isObject()) {
            metadata = doc.object();
        }
        metadataFile.close();
    }
    
    // If no JSON metadata, try to extract from binary
    if (metadata.isEmpty()) {
        metadata = extractMetadataFromBinary(pluginPath);
    }
    
    return metadata;
}

QJsonObject PluginManager::extractMetadataFromBinary(const QString &pluginPath)
{
    QJsonObject metadata;
    QFile file(pluginPath);
    
    if (!file.open(QIODevice::ReadOnly)) {
        return metadata;
    }
    
    // Read file content to look for metadata strings
    QByteArray content = file.readAll();
    file.close();
    
    // Look for common metadata patterns
    QString contentStr = QString::fromUtf8(content);
    
    // Extract version
    QRegularExpression versionRegex("version[\\s]*[:=][\\s]*[\"']([0-9]+\\.[0-9]+\\.[0-9]+)[\"']");
    QRegularExpressionMatch versionMatch = versionRegex.match(contentStr);
    if (versionMatch.hasMatch()) {
        metadata["version"] = versionMatch.captured(1);
    }
    
    // Extract description
    QRegularExpression descRegex("description[\\s]*[:=][\\s]*[\"']([^\"]+)[\"']");
    QRegularExpressionMatch descMatch = descRegex.match(contentStr);
    if (descMatch.hasMatch()) {
        metadata["description"] = descMatch.captured(1);
    }
    
    // Extract author
    QRegularExpression authorRegex("author[\\s]*[:=][\\s]*[\"']([^\"]+)[\"']");
    QRegularExpressionMatch authorMatch = authorRegex.match(contentStr);
    if (authorMatch.hasMatch()) {
        metadata["author"] = authorMatch.captured(1);
    }
    
    // Extract license
    QRegularExpression licenseRegex("license[\\s]*[:=][\\s]*[\"']([^\"]+)[\"']");
    QRegularExpressionMatch licenseMatch = licenseRegex.match(contentStr);
    if (licenseMatch.hasMatch()) {
        metadata["license"] = licenseMatch.captured(1);
    }
    
    return metadata;
}

bool PluginManager::validatePluginFile(const QString &pluginPath)
{
    // In a real implementation, this would validate the plugin file
    // Check if it's a valid shared library and contains required symbols
    
    QFileInfo fileInfo(pluginPath);
    if (!fileInfo.exists() || !fileInfo.isFile()) {
        return false;
    }
    
    // Check file extension
    QString extension = fileInfo.suffix().toLower();
#ifdef Q_OS_WIN
    if (extension != "dll") return false;
#elif defined(Q_OS_MAC)
    if (extension != "dylib") return false;
#else
    if (extension != "so") return false;
#endif
    
    return true;
}

bool PluginManager::installPlugin(const QString &pluginPath)
{
    QString pluginsDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/plugins";
    QDir dir(pluginsDir);
    
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    
    QFileInfo sourceInfo(pluginPath);
    QString destinationPath = dir.absoluteFilePath(sourceInfo.fileName());
    
    // Copy plugin file
    QFile sourceFile(pluginPath);
    if (!sourceFile.copy(destinationPath)) {
        qWarning() << "Failed to copy plugin file:" << sourceFile.errorString();
        return false;
    }
    
    // Set executable permissions
    QFile destFile(destinationPath);
    destFile.setPermissions(destFile.permissions() | QFile::ExeOwner | QFile::ExeUser);
    
    return true;
}
