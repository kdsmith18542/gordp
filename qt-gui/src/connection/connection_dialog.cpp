#include "connection_dialog.h"
#include "ui_connection_dialog.h"
#include <QMessageBox>
#include <QJsonDocument>
#include <QApplication>
#include <QProgressDialog>
#include <QTimer>
#include <QStandardPaths>
#include <QDebug>
#include "../utils/gordp_bridge.h"
#include "../history/history_dialog.h"
#include "../favorites/favorites_dialog.h"

ConnectionDialog::ConnectionDialog(QWidget *parent)
    : QDialog(parent)
    , m_serverEdit(new QLineEdit(this))
    , m_portSpinBox(new QSpinBox(this))
    , m_usernameEdit(new QLineEdit(this))
    , m_passwordEdit(new QLineEdit(this))
    , m_savePasswordCheckBox(new QCheckBox("Save password", this))
    , m_colorDepthComboBox(new QComboBox(this))
    , m_resolutionComboBox(new QComboBox(this))
    , m_fullscreenCheckBox(new QCheckBox("Fullscreen", this))
    , m_audioCheckBox(new QCheckBox("Enable audio", this))
    , m_clipboardCheckBox(new QCheckBox("Enable clipboard", this))
    , m_driveRedirectionCheckBox(new QCheckBox("Enable drive redirection", this))
    , m_connectButton(new QPushButton("Connect", this))
    , m_cancelButton(new QPushButton("Cancel", this))
    , m_testButton(new QPushButton("Test Connection", this))
    , m_historyButton(new QPushButton("History", this))
    , m_favoritesButton(new QPushButton("Favorites", this))
{
    setWindowTitle("Connect to Remote Server");
    setModal(true);
    setMinimumSize(400, 500);
    
    setupUI();
    loadSettings();
    
    // Connect signals
    connect(m_connectButton, &QPushButton::clicked, this, &ConnectionDialog::onConnectClicked);
    connect(m_cancelButton, &QPushButton::clicked, this, &ConnectionDialog::onCancelClicked);
    connect(m_testButton, &QPushButton::clicked, this, &ConnectionDialog::onTestConnectionClicked);
    connect(m_historyButton, &QPushButton::clicked, this, &ConnectionDialog::onLoadFromHistory);
    connect(m_favoritesButton, &QPushButton::clicked, this, &ConnectionDialog::onSaveToFavorites);
    
    // Update connect button state
    connect(m_serverEdit, &QLineEdit::textChanged, this, &ConnectionDialog::updateConnectButton);
    connect(m_usernameEdit, &QLineEdit::textChanged, this, &ConnectionDialog::updateConnectButton);
    
    updateConnectButton();
}

ConnectionDialog::~ConnectionDialog()
{
    saveSettings();
}

void ConnectionDialog::setupUI()
{
    QVBoxLayout *mainLayout = new QVBoxLayout(this);
    
    // Server settings group
    QGroupBox *serverGroup = new QGroupBox("Server Settings", this);
    QFormLayout *serverLayout = new QFormLayout(serverGroup);
    
    m_serverEdit->setPlaceholderText("Enter server address (e.g., 192.168.1.100)");
    serverLayout->addRow("Server:", m_serverEdit);
    
    m_portSpinBox->setRange(1, 65535);
    m_portSpinBox->setValue(3389);
    serverLayout->addRow("Port:", m_portSpinBox);
    
    mainLayout->addWidget(serverGroup);
    
    // Authentication group
    QGroupBox *authGroup = new QGroupBox("Authentication", this);
    QFormLayout *authLayout = new QFormLayout(authGroup);
    
    m_usernameEdit->setPlaceholderText("Enter username");
    authLayout->addRow("Username:", m_usernameEdit);
    
    m_passwordEdit->setEchoMode(QLineEdit::Password);
    m_passwordEdit->setPlaceholderText("Enter password");
    authLayout->addRow("Password:", m_passwordEdit);
    
    authLayout->addRow("", m_savePasswordCheckBox);
    
    mainLayout->addWidget(authGroup);
    
    // Display settings group
    QGroupBox *displayGroup = new QGroupBox("Display Settings", this);
    QFormLayout *displayLayout = new QFormLayout(displayGroup);
    
    m_colorDepthComboBox->addItems({"16-bit", "24-bit", "32-bit"});
    m_colorDepthComboBox->setCurrentText("24-bit");
    displayLayout->addRow("Color Depth:", m_colorDepthComboBox);
    
    m_resolutionComboBox->addItems({
        "1024x768", "1280x720", "1280x1024", "1366x768", 
        "1440x900", "1600x900", "1920x1080", "Full Screen"
    });
    m_resolutionComboBox->setCurrentText("1024x768");
    displayLayout->addRow("Resolution:", m_resolutionComboBox);
    
    displayLayout->addRow("", m_fullscreenCheckBox);
    
    mainLayout->addWidget(displayGroup);
    
    // Features group
    QGroupBox *featuresGroup = new QGroupBox("Features", this);
    QVBoxLayout *featuresLayout = new QVBoxLayout(featuresGroup);
    
    featuresLayout->addWidget(m_audioCheckBox);
    featuresLayout->addWidget(m_clipboardCheckBox);
    featuresLayout->addWidget(m_driveRedirectionCheckBox);
    
    mainLayout->addWidget(featuresGroup);
    
    // Buttons
    QHBoxLayout *buttonLayout = new QHBoxLayout();
    
    buttonLayout->addWidget(m_historyButton);
    buttonLayout->addWidget(m_favoritesButton);
    buttonLayout->addWidget(m_testButton);
    buttonLayout->addStretch();
    buttonLayout->addWidget(m_cancelButton);
    buttonLayout->addWidget(m_connectButton);
    
    mainLayout->addLayout(buttonLayout);
    
    // Set default focus
    m_serverEdit->setFocus();
}

void ConnectionDialog::loadSettings()
{
    QSettings settings;
    
    // Load last used values
    m_serverEdit->setText(settings.value("Connection/LastServer", "").toString());
    m_portSpinBox->setValue(settings.value("Connection/LastPort", 3389).toInt());
    m_usernameEdit->setText(settings.value("Connection/LastUsername", "").toString());
    
    // Load saved password if enabled
    if (settings.value("Connection/SavePassword", false).toBool()) {
        m_passwordEdit->setText(settings.value("Connection/LastPassword", "").toString());
        m_savePasswordCheckBox->setChecked(true);
    }
    
    // Load display settings
    m_colorDepthComboBox->setCurrentText(settings.value("Connection/ColorDepth", "24-bit").toString());
    m_resolutionComboBox->setCurrentText(settings.value("Connection/Resolution", "1024x768").toString());
    m_fullscreenCheckBox->setChecked(settings.value("Connection/Fullscreen", false).toBool());
    
    // Load feature settings
    m_audioCheckBox->setChecked(settings.value("Connection/Audio", true).toBool());
    m_clipboardCheckBox->setChecked(settings.value("Connection/Clipboard", true).toBool());
    m_driveRedirectionCheckBox->setChecked(settings.value("Connection/DriveRedirection", false).toBool());
}

void ConnectionDialog::saveSettings()
{
    QSettings settings;
    
    // Save current values
    settings.setValue("Connection/LastServer", m_serverEdit->text());
    settings.setValue("Connection/LastPort", m_portSpinBox->value());
    settings.setValue("Connection/LastUsername", m_usernameEdit->text());
    
    // Save password if enabled
    if (m_savePasswordCheckBox->isChecked()) {
        settings.setValue("Connection/LastPassword", m_passwordEdit->text());
        settings.setValue("Connection/SavePassword", true);
    } else {
        settings.remove("Connection/LastPassword");
        settings.setValue("Connection/SavePassword", false);
    }
    
    // Save display settings
    settings.setValue("Connection/ColorDepth", m_colorDepthComboBox->currentText());
    settings.setValue("Connection/Resolution", m_resolutionComboBox->currentText());
    settings.setValue("Connection/Fullscreen", m_fullscreenCheckBox->isChecked());
    
    // Save feature settings
    settings.setValue("Connection/Audio", m_audioCheckBox->isChecked());
    settings.setValue("Connection/Clipboard", m_clipboardCheckBox->isChecked());
    settings.setValue("Connection/DriveRedirection", m_driveRedirectionCheckBox->isChecked());
}

void ConnectionDialog::updateConnectButton()
{
    bool canConnect = !m_serverEdit->text().isEmpty() && !m_usernameEdit->text().isEmpty();
    m_connectButton->setEnabled(canConnect);
}

QJsonObject ConnectionDialog::getConnectionOptions() const
{
    QJsonObject options;
    
    // Display options
    options["colorDepth"] = m_colorDepthComboBox->currentText();
    options["resolution"] = m_resolutionComboBox->currentText();
    options["fullscreen"] = m_fullscreenCheckBox->isChecked();
    
    // Feature options
    options["audio"] = m_audioCheckBox->isChecked();
    options["clipboard"] = m_clipboardCheckBox->isChecked();
    options["driveRedirection"] = m_driveRedirectionCheckBox->isChecked();
    
    return options;
}

void ConnectionDialog::onConnectClicked()
{
    QString server = m_serverEdit->text().trimmed();
    int port = m_portSpinBox->value();
    QString username = m_usernameEdit->text().trimmed();
    QString password = m_passwordEdit->text();
    
    if (server.isEmpty()) {
        QMessageBox::warning(this, "Error", "Please enter a server address.");
        m_serverEdit->setFocus();
        return;
    }
    
    if (username.isEmpty()) {
        QMessageBox::warning(this, "Error", "Please enter a username.");
        m_usernameEdit->setFocus();
        return;
    }
    
    // Save settings
    saveSettings();
    
    // Emit connection request
    emit connectRequested(server, port, username, password, getConnectionOptions());
    
    // Close dialog
    accept();
}

void ConnectionDialog::onCancelClicked()
{
    reject();
}

void ConnectionDialog::onTestConnectionClicked()
{
    QString server = m_serverEdit->text().trimmed();
    int port = m_portSpinBox->value();
    
    if (server.isEmpty()) {
        QMessageBox::warning(this, "Error", "Please enter a server address.");
        return;
    }
    
    // Show progress dialog
    QProgressDialog progress("Testing connection...", "Cancel", 0, 100, this);
    progress.setWindowModality(Qt::WindowModal);
    progress.setAutoClose(false);
    progress.show();
    
    // Create temporary GoRDP bridge for testing
    GoRDPBridge *testBridge = new GoRDPBridge(this);
    
    // Connect signals for test results
    connect(testBridge, &GoRDPBridge::connectionStatusChanged, this, [&](bool connected) {
        progress.close();
        if (connected) {
            QMessageBox::information(this, "Test Connection", 
                QString("Successfully connected to %1:%2").arg(server).arg(port));
        } else {
            QMessageBox::warning(this, "Test Connection", 
                QString("Failed to connect to %1:%2\n\nPlease check:\n• Server address and port\n• Network connectivity\n• Firewall settings\n• RDP service on server").arg(server).arg(port));
        }
        testBridge->disconnectFromServer();
        testBridge->deleteLater();
    });
    
    connect(testBridge, &GoRDPBridge::errorOccurred, this, [&](const QString &error) {
        progress.close();
        QMessageBox::critical(this, "Test Connection Error", 
            QString("Error testing connection to %1:%2\n\n%3").arg(server).arg(port).arg(error));
        testBridge->deleteLater();
    });
    
    // Start connection test
    progress.setValue(25);
    QApplication::processEvents();
    
    // Test connection with minimal options
    QJsonObject testOptions;
    testOptions["testMode"] = true;
    testOptions["timeout"] = 10000; // 10 second timeout
    
    testBridge->connectToServer(server, port, "", "", testOptions);
    
    progress.setValue(50);
    QApplication::processEvents();
    
    // Set a timeout
    QTimer::singleShot(15000, this, [&]() {
        if (progress.isVisible()) {
            progress.close();
            QMessageBox::warning(this, "Test Connection", 
                QString("Connection test to %1:%2 timed out").arg(server).arg(port));
            testBridge->disconnectFromServer();
            testBridge->deleteLater();
        }
    });
}

void ConnectionDialog::onLoadFromHistory()
{
    // Create history dialog if not exists
    if (!m_historyDialog) {
        m_historyDialog = new HistoryDialog(this);
        connect(m_historyDialog, &HistoryDialog::connectionSelected, this, [this](const QJsonObject &connection) {
            // Load connection details into form
            m_serverEdit->setText(connection["server"].toString());
            m_portSpinBox->setValue(connection["port"].toInt());
            m_usernameEdit->setText(connection["username"].toString());
            
            // Load display settings if available
            if (connection.contains("colorDepth")) {
                m_colorDepthComboBox->setCurrentText(connection["colorDepth"].toString());
            }
            if (connection.contains("resolution")) {
                m_resolutionComboBox->setCurrentText(connection["resolution"].toString());
            }
            if (connection.contains("fullscreen")) {
                m_fullscreenCheckBox->setChecked(connection["fullscreen"].toBool());
            }
            
            // Load feature settings if available
            if (connection.contains("audio")) {
                m_audioCheckBox->setChecked(connection["audio"].toBool());
            }
            if (connection.contains("clipboard")) {
                m_clipboardCheckBox->setChecked(connection["clipboard"].toBool());
            }
            if (connection.contains("driveRedirection")) {
                m_driveRedirectionCheckBox->setChecked(connection["driveRedirection"].toBool());
            }
            
            m_historyDialog->close();
        });
    }
    
    m_historyDialog->show();
    m_historyDialog->raise();
    m_historyDialog->activateWindow();
}

void ConnectionDialog::onSaveToFavorites()
{
    QString server = m_serverEdit->text().trimmed();
    QString username = m_usernameEdit->text().trimmed();
    
    if (server.isEmpty()) {
        QMessageBox::warning(this, "Error", "Please enter a server address.");
        return;
    }
    
    // Create favorites dialog if not exists
    if (!m_favoritesDialog) {
        m_favoritesDialog = new FavoritesDialog(this);
    }
    
    // Prepare favorite data
    QJsonObject favorite;
    favorite["name"] = QString("%1@%2").arg(username).arg(server);
    favorite["server"] = server;
    favorite["port"] = m_portSpinBox->value();
    favorite["username"] = username;
    favorite["colorDepth"] = m_colorDepthComboBox->currentText();
    favorite["resolution"] = m_resolutionComboBox->currentText();
    favorite["fullscreen"] = m_fullscreenCheckBox->isChecked();
    favorite["audio"] = m_audioCheckBox->isChecked();
    favorite["clipboard"] = m_clipboardCheckBox->isChecked();
    favorite["driveRedirection"] = m_driveRedirectionCheckBox->isChecked();
    
    // Add to favorites
    m_favoritesDialog->addFavorite(favorite);
    
    QMessageBox::information(this, "Save to Favorites", 
        QString("Successfully saved %1@%2 to favorites.").arg(username).arg(server));
} 