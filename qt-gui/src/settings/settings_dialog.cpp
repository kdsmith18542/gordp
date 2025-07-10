#include "settings_dialog.h"
#include "ui_settings_dialog.h"
#include <QMessageBox>
#include <QJsonDocument>

SettingsDialog::SettingsDialog(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::SettingsDialog)
    , m_settings(new QSettings("GoRDP", "GUI"))
    , m_settingsModified(false)
{
    ui->setupUi(this);
    setupConnections();
    loadSettings();
}

SettingsDialog::~SettingsDialog()
{
    delete ui;
    delete m_settings;
}

void SettingsDialog::loadSettings()
{
    // General settings
    ui->startMinimizedCheckBox->setChecked(m_settings->value("startup/minimized", false).toBool());
    ui->autoConnectCheckBox->setChecked(m_settings->value("startup/autoConnect", false).toBool());
    ui->checkUpdatesCheckBox->setChecked(m_settings->value("startup/checkUpdates", true).toBool());
    
    // Display settings
    int colorDepth = m_settings->value("display/colorDepth", 1).toInt();
    ui->defaultColorDepthComboBox->setCurrentIndex(colorDepth);
    
    int resolution = m_settings->value("display/resolution", 0).toInt();
    ui->defaultResolutionComboBox->setCurrentIndex(resolution);
    
    // Security settings
    ui->enableEncryptionCheckBox->setChecked(m_settings->value("security/enableEncryption", true).toBool());
    ui->requireEncryptionCheckBox->setChecked(m_settings->value("security/requireEncryption", false).toBool());
    ui->enableNLA->setChecked(m_settings->value("security/enableNLA", true).toBool());
    ui->verifyCertificatesCheckBox->setChecked(m_settings->value("security/verifyCertificates", true).toBool());
    ui->warnOnCertMismatchCheckBox->setChecked(m_settings->value("security/warnOnCertMismatch", true).toBool());
    
    // Performance settings
    ui->enableHardwareAccelerationCheckBox->setChecked(m_settings->value("performance/hardwareAcceleration", true).toBool());
    ui->enableBitmapCachingCheckBox->setChecked(m_settings->value("performance/bitmapCaching", true).toBool());
    ui->enableCompressionCheckBox->setChecked(m_settings->value("performance/compression", true).toBool());
    
    int imageQuality = m_settings->value("performance/imageQuality", 1).toInt();
    ui->imageQualityComboBox->setCurrentIndex(imageQuality);
    
    m_settingsModified = false;
    updateUI();
}

void SettingsDialog::saveSettings()
{
    // General settings
    m_settings->setValue("startup/minimized", ui->startMinimizedCheckBox->isChecked());
    m_settings->setValue("startup/autoConnect", ui->autoConnectCheckBox->isChecked());
    m_settings->setValue("startup/checkUpdates", ui->checkUpdatesCheckBox->isChecked());
    
    // Display settings
    m_settings->setValue("display/colorDepth", ui->defaultColorDepthComboBox->currentIndex());
    m_settings->setValue("display/resolution", ui->defaultResolutionComboBox->currentIndex());
    
    // Security settings
    m_settings->setValue("security/enableEncryption", ui->enableEncryptionCheckBox->isChecked());
    m_settings->setValue("security/requireEncryption", ui->requireEncryptionCheckBox->isChecked());
    m_settings->setValue("security/enableNLA", ui->enableNLA->isChecked());
    m_settings->setValue("security/verifyCertificates", ui->verifyCertificatesCheckBox->isChecked());
    m_settings->setValue("security/warnOnCertMismatch", ui->warnOnCertMismatchCheckBox->isChecked());
    
    // Performance settings
    m_settings->setValue("performance/hardwareAcceleration", ui->enableHardwareAccelerationCheckBox->isChecked());
    m_settings->setValue("performance/bitmapCaching", ui->enableBitmapCachingCheckBox->isChecked());
    m_settings->setValue("performance/compression", ui->enableCompressionCheckBox->isChecked());
    m_settings->setValue("performance/imageQuality", ui->imageQualityComboBox->currentIndex());
    
    m_settings->sync();
    m_settingsModified = false;
    updateUI();
    
    // Create JSON settings object
    m_currentSettings["startup"] = QJsonObject{
        {"minimized", ui->startMinimizedCheckBox->isChecked()},
        {"autoConnect", ui->autoConnectCheckBox->isChecked()},
        {"checkUpdates", ui->checkUpdatesCheckBox->isChecked()}
    };
    
    m_currentSettings["display"] = QJsonObject{
        {"colorDepth", ui->defaultColorDepthComboBox->currentIndex()},
        {"resolution", ui->defaultResolutionComboBox->currentIndex()}
    };
    
    m_currentSettings["security"] = QJsonObject{
        {"enableEncryption", ui->enableEncryptionCheckBox->isChecked()},
        {"requireEncryption", ui->requireEncryptionCheckBox->isChecked()},
        {"enableNLA", ui->enableNLA->isChecked()},
        {"verifyCertificates", ui->verifyCertificatesCheckBox->isChecked()},
        {"warnOnCertMismatch", ui->warnOnCertMismatchCheckBox->isChecked()}
    };
    
    m_currentSettings["performance"] = QJsonObject{
        {"hardwareAcceleration", ui->enableHardwareAccelerationCheckBox->isChecked()},
        {"bitmapCaching", ui->enableBitmapCachingCheckBox->isChecked()},
        {"compression", ui->enableCompressionCheckBox->isChecked()},
        {"imageQuality", ui->imageQualityComboBox->currentIndex()}
    };
    
    emit settingsChanged(m_currentSettings);
}

void SettingsDialog::resetToDefaults()
{
    int result = QMessageBox::question(this, "Reset Settings", 
                                      "Are you sure you want to reset all settings to defaults?",
                                      QMessageBox::Yes | QMessageBox::No);
    
    if (result == QMessageBox::Yes) {
        m_settings->clear();
        loadSettings();
        m_settingsModified = true;
        updateUI();
    }
}

QJsonObject SettingsDialog::getSettings() const
{
    return m_currentSettings;
}

void SettingsDialog::onOkClicked()
{
    saveSettings();
    accept();
}

void SettingsDialog::onCancelClicked()
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

void SettingsDialog::onResetClicked()
{
    resetToDefaults();
}

void SettingsDialog::onSettingsChanged()
{
    m_settingsModified = true;
    updateUI();
}

void SettingsDialog::setupConnections()
{
    connect(ui->okButton, &QPushButton::clicked, this, &SettingsDialog::onOkClicked);
    connect(ui->cancelButton, &QPushButton::clicked, this, &SettingsDialog::onCancelClicked);
    connect(ui->resetButton, &QPushButton::clicked, this, &SettingsDialog::onResetClicked);
    
    // Connect all checkboxes and comboboxes to settings changed
    connect(ui->startMinimizedCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->autoConnectCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->checkUpdatesCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->defaultColorDepthComboBox, QOverload<int>::of(&QComboBox::currentIndexChanged), 
            this, &SettingsDialog::onSettingsChanged);
    connect(ui->defaultResolutionComboBox, QOverload<int>::of(&QComboBox::currentIndexChanged), 
            this, &SettingsDialog::onSettingsChanged);
    connect(ui->enableEncryptionCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->requireEncryptionCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->enableNLA, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->verifyCertificatesCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->warnOnCertMismatchCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->enableHardwareAccelerationCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->enableBitmapCachingCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->enableCompressionCheckBox, &QCheckBox::toggled, this, &SettingsDialog::onSettingsChanged);
    connect(ui->imageQualityComboBox, QOverload<int>::of(&QComboBox::currentIndexChanged), 
            this, &SettingsDialog::onSettingsChanged);
}

void SettingsDialog::updateUI()
{
    ui->okButton->setEnabled(m_settingsModified);
    ui->resetButton->setEnabled(true);
} 