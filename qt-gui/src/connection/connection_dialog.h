#ifndef CONNECTION_DIALOG_H
#define CONNECTION_DIALOG_H

#include <QDialog>
#include <QLineEdit>
#include <QSpinBox>
#include <QCheckBox>
#include <QComboBox>
#include <QPushButton>
#include <QGroupBox>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QFormLayout>
#include <QLabel>
#include <QJsonObject>
#include <QSettings>
#include <QProgressDialog>
#include <QTimer>

// Forward declarations
class HistoryDialog;
class FavoritesDialog;
class GoRDPBridge;

class ConnectionDialog : public QDialog
{
    Q_OBJECT

public:
    explicit ConnectionDialog(QWidget *parent = nullptr);
    ~ConnectionDialog();

signals:
    void connectRequested(const QString &server, int port, 
                         const QString &username, const QString &password,
                         const QJsonObject &options);

private slots:
    void onConnectClicked();
    void onCancelClicked();
    void onTestConnectionClicked();
    void onLoadFromHistory();
    void onSaveToFavorites();

private:
    void setupUI();
    void loadSettings();
    void saveSettings();
    void updateConnectButton();
    QJsonObject getConnectionOptions() const;

    // UI Components
    QLineEdit *m_serverEdit;
    QSpinBox *m_portSpinBox;
    QLineEdit *m_usernameEdit;
    QLineEdit *m_passwordEdit;
    QCheckBox *m_savePasswordCheckBox;
    QComboBox *m_colorDepthComboBox;
    QComboBox *m_resolutionComboBox;
    QCheckBox *m_fullscreenCheckBox;
    QCheckBox *m_audioCheckBox;
    QCheckBox *m_clipboardCheckBox;
    QCheckBox *m_driveRedirectionCheckBox;
    QPushButton *m_connectButton;
    QPushButton *m_cancelButton;
    QPushButton *m_testButton;
    QPushButton *m_historyButton;
    QPushButton *m_favoritesButton;
    
    // Additional members
    QSettings *m_settings;
    HistoryDialog *m_historyDialog;
    FavoritesDialog *m_favoritesDialog;
    QVariantList m_recentConnections;
};

#endif // CONNECTION_DIALOG_H 