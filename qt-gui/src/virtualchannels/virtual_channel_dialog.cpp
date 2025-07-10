#include "virtual_channel_dialog.h"
#include "ui_virtual_channel_dialog.h"
#include <QMessageBox>
#include <QJsonDocument>
#include <QApplication>
#include <QClipboard>

VirtualChannelDialog::VirtualChannelDialog(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::VirtualChannelDialog)
    , m_settings(new QSettings("GoRDP", "VirtualChannels"))
    , m_settingsModified(false)
    , m_clipboardEnabled(false)
    , m_audioEnabled(false)
    , m_deviceRedirectionEnabled(false)
{
    ui->setupUi(this);
    setupConnections();
    loadSettings();
}

VirtualChannelDialog::~VirtualChannelDialog()
{
    delete ui;
    delete m_settings;
}

void VirtualChannelDialog::loadSettings()
{
    // Clipboard settings
    ui->enableClipboardCheckBox->setChecked(m_settings->value("clipboard/enabled", true).toBool());
    ui->clipboardTextCheckBox->setChecked(m_settings->value("clipboard/text", true).toBool());
    ui->clipboardImageCheckBox->setChecked(m_settings->value("clipboard/images", true).toBool());
    ui->clipboardFileCheckBox->setChecked(m_settings->value("clipboard/files", false).toBool());
    
    // Audio settings
    ui->enableAudioCheckBox->setChecked(m_settings->value("audio/enabled", true).toBool());
    ui->audioPlaybackCheckBox->setChecked(m_settings->value("audio/playback", true).toBool());
    ui->audioRecordingCheckBox->setChecked(m_settings->value("audio/recording", false).toBool());
    
    // Device redirection settings
    ui->enableDriveRedirectionCheckBox->setChecked(m_settings->value("devices/drive", false).toBool());
    ui->enablePrinterRedirectionCheckBox->setChecked(m_settings->value("devices/printer", false).toBool());
    ui->enablePortRedirectionCheckBox->setChecked(m_settings->value("devices/port", false).toBool());
    
    m_settingsModified = false;
    updateUI();
    updateStatusLabels();
}

void VirtualChannelDialog::saveSettings()
{
    // Clipboard settings
    m_settings->setValue("clipboard/enabled", ui->enableClipboardCheckBox->isChecked());
    m_settings->setValue("clipboard/text", ui->clipboardTextCheckBox->isChecked());
    m_settings->setValue("clipboard/images", ui->clipboardImageCheckBox->isChecked());
    m_settings->setValue("clipboard/files", ui->clipboardFileCheckBox->isChecked());
    
    // Audio settings
    m_settings->setValue("audio/enabled", ui->enableAudioCheckBox->isChecked());
    m_settings->setValue("audio/playback", ui->audioPlaybackCheckBox->isChecked());
    m_settings->setValue("audio/recording", ui->audioRecordingCheckBox->isChecked());
    
    // Device redirection settings
    m_settings->setValue("devices/drive", ui->enableDriveRedirectionCheckBox->isChecked());
    m_settings->setValue("devices/printer", ui->enablePrinterRedirectionCheckBox->isChecked());
    m_settings->setValue("devices/port", ui->enablePortRedirectionCheckBox->isChecked());
    
    m_settings->sync();
    m_settingsModified = false;
    updateUI();
    
    // Create JSON settings object
    m_currentSettings["clipboard"] = QJsonObject{
        {"enabled", ui->enableClipboardCheckBox->isChecked()},
        {"text", ui->clipboardTextCheckBox->isChecked()},
        {"images", ui->clipboardImageCheckBox->isChecked()},
        {"files", ui->clipboardFileCheckBox->isChecked()}
    };
    
    m_currentSettings["audio"] = QJsonObject{
        {"enabled", ui->enableAudioCheckBox->isChecked()},
        {"playback", ui->audioPlaybackCheckBox->isChecked()},
        {"recording", ui->audioRecordingCheckBox->isChecked()}
    };
    
    m_currentSettings["devices"] = QJsonObject{
        {"drive", ui->enableDriveRedirectionCheckBox->isChecked()},
        {"printer", ui->enablePrinterRedirectionCheckBox->isChecked()},
        {"port", ui->enablePortRedirectionCheckBox->isChecked()}
    };
    
    emit settingsChanged(m_currentSettings);
}

QJsonObject VirtualChannelDialog::getChannelSettings() const
{
    return m_currentSettings;
}

void VirtualChannelDialog::updateClipboardStatus(bool enabled, const QString &status)
{
    m_clipboardEnabled = enabled;
    m_clipboardStatus = status;
    updateStatusLabels();
}

void VirtualChannelDialog::updateAudioStatus(bool enabled, const QString &status)
{
    m_audioEnabled = enabled;
    m_audioStatus = status;
    updateStatusLabels();
}

void VirtualChannelDialog::updateDeviceStatus(bool enabled, const QString &status)
{
    m_deviceRedirectionEnabled = enabled;
    m_deviceStatus = status;
    updateStatusLabels();
}

void VirtualChannelDialog::onApplyClicked()
{
    saveSettings();
    
    // Apply settings to active connection
    emit clipboardToggled(ui->enableClipboardCheckBox->isChecked());
    emit audioToggled(ui->enableAudioCheckBox->isChecked());
    emit deviceRedirectionToggled(ui->enableDriveRedirectionCheckBox->isChecked() || 
                                 ui->enablePrinterRedirectionCheckBox->isChecked() || 
                                 ui->enablePortRedirectionCheckBox->isChecked());
    
    QMessageBox::information(this, "Settings Applied", 
                           "Virtual channel settings have been applied to the current connection.");
}

void VirtualChannelDialog::onCloseClicked()
{
    if (m_settingsModified) {
        int result = QMessageBox::question(this, "Unsaved Changes", 
                                          "You have unsaved changes. Do you want to save them?",
                                          QMessageBox::Yes | QMessageBox::No | QMessageBox::Cancel);
        
        if (result == QMessageBox::Yes) {
            saveSettings();
            accept();
        } else if (result == QMessageBox::No) {
            reject();
        }
        // Cancel - do nothing, stay in dialog
    } else {
        reject();
    }
}

void VirtualChannelDialog::onClipboardToggled(bool enabled)
{
    m_settingsModified = true;
    updateUI();
    
    // Enable/disable clipboard sub-options
    ui->clipboardTextCheckBox->setEnabled(enabled);
    ui->clipboardImageCheckBox->setEnabled(enabled);
    ui->clipboardFileCheckBox->setEnabled(enabled);
    
    emit clipboardToggled(enabled);
}

void VirtualChannelDialog::onAudioToggled(bool enabled)
{
    m_settingsModified = true;
    updateUI();
    
    // Enable/disable audio sub-options
    ui->audioPlaybackCheckBox->setEnabled(enabled);
    ui->audioRecordingCheckBox->setEnabled(enabled);
    
    emit audioToggled(enabled);
}

void VirtualChannelDialog::onDeviceRedirectionToggled(bool enabled)
{
    m_settingsModified = true;
    updateUI();
    
    // Note: Individual device options are handled separately
    // This is just for the overall device redirection status
}

void VirtualChannelDialog::onSettingsChanged()
{
    m_settingsModified = true;
    updateUI();
}

void VirtualChannelDialog::setupConnections()
{
    connect(ui->applyButton, &QPushButton::clicked, this, &VirtualChannelDialog::onApplyClicked);
    connect(ui->closeButton, &QPushButton::clicked, this, &VirtualChannelDialog::onCloseClicked);
    
    // Connect clipboard settings
    connect(ui->enableClipboardCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onClipboardToggled);
    connect(ui->clipboardTextCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    connect(ui->clipboardImageCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    connect(ui->clipboardFileCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    
    // Connect audio settings
    connect(ui->enableAudioCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onAudioToggled);
    connect(ui->audioPlaybackCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    connect(ui->audioRecordingCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    
    // Connect device redirection settings
    connect(ui->enableDriveRedirectionCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    connect(ui->enablePrinterRedirectionCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
    connect(ui->enablePortRedirectionCheckBox, &QCheckBox::toggled, this, &VirtualChannelDialog::onSettingsChanged);
}

void VirtualChannelDialog::updateUI()
{
    ui->applyButton->setEnabled(m_settingsModified);
    
    // Update clipboard sub-options state
    bool clipboardEnabled = ui->enableClipboardCheckBox->isChecked();
    ui->clipboardTextCheckBox->setEnabled(clipboardEnabled);
    ui->clipboardImageCheckBox->setEnabled(clipboardEnabled);
    ui->clipboardFileCheckBox->setEnabled(clipboardEnabled);
    
    // Update audio sub-options state
    bool audioEnabled = ui->enableAudioCheckBox->isChecked();
    ui->audioPlaybackCheckBox->setEnabled(audioEnabled);
    ui->audioRecordingCheckBox->setEnabled(audioEnabled);
}

void VirtualChannelDialog::updateStatusLabels()
{
    // Update status labels with current connection status
    QString clipboardText = m_clipboardEnabled ? 
        QString("Active - %1").arg(m_clipboardStatus) : "Inactive";
    QString audioText = m_audioEnabled ? 
        QString("Active - %1").arg(m_audioStatus) : "Inactive";
    QString deviceText = m_deviceRedirectionEnabled ? 
        QString("Active - %1").arg(m_deviceStatus) : "Inactive";
    
    // Note: In a real implementation, you would have QLabel widgets in the UI
    // to display these status messages. For now, we'll use qDebug for demonstration.
    qDebug() << "Clipboard:" << clipboardText;
    qDebug() << "Audio:" << audioText;
    qDebug() << "Device Redirection:" << deviceText;
}
