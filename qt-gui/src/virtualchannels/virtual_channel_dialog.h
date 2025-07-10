#ifndef VIRTUAL_CHANNEL_DIALOG_H
#define VIRTUAL_CHANNEL_DIALOG_H

#include <QDialog>
#include <QJsonObject>
#include <QSettings>

namespace Ui {
class VirtualChannelDialog;
}

class VirtualChannelDialog : public QDialog
{
    Q_OBJECT

public:
    explicit VirtualChannelDialog(QWidget *parent = nullptr);
    ~VirtualChannelDialog();

    // Virtual channel management
    void loadSettings();
    void saveSettings();
    QJsonObject getChannelSettings() const;
    
    // Channel status
    void updateClipboardStatus(bool enabled, const QString &status);
    void updateAudioStatus(bool enabled, const QString &status);
    void updateDeviceStatus(bool enabled, const QString &status);

signals:
    void settingsChanged(const QJsonObject &settings);
    void clipboardToggled(bool enabled);
    void audioToggled(bool enabled);
    void deviceRedirectionToggled(bool enabled);

private slots:
    void onApplyClicked();
    void onCloseClicked();
    void onClipboardToggled(bool enabled);
    void onAudioToggled(bool enabled);
    void onDeviceRedirectionToggled(bool enabled);
    void onSettingsChanged();

private:
    void setupConnections();
    void updateUI();
    void updateStatusLabels();
    
    Ui::VirtualChannelDialog *ui;
    QSettings *m_settings;
    QJsonObject m_currentSettings;
    bool m_settingsModified;
    
    // Status tracking
    bool m_clipboardEnabled;
    bool m_audioEnabled;
    bool m_deviceRedirectionEnabled;
    QString m_clipboardStatus;
    QString m_audioStatus;
    QString m_deviceStatus;
};

#endif // VIRTUAL_CHANNEL_DIALOG_H 