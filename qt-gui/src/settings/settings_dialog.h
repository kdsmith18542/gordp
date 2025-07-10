#ifndef SETTINGS_DIALOG_H
#define SETTINGS_DIALOG_H

#include <QDialog>
#include <QSettings>
#include <QJsonObject>

namespace Ui {
class SettingsDialog;
}

class SettingsDialog : public QDialog
{
    Q_OBJECT

public:
    explicit SettingsDialog(QWidget *parent = nullptr);
    ~SettingsDialog();

    // Settings management
    void loadSettings();
    void saveSettings();
    void resetToDefaults();
    
    // Get settings
    QJsonObject getSettings() const;

signals:
    void settingsChanged(const QJsonObject &settings);

private slots:
    void onOkClicked();
    void onCancelClicked();
    void onResetClicked();
    void onSettingsChanged();

private:
    void setupConnections();
    void updateUI();
    
    Ui::SettingsDialog *ui;
    QSettings *m_settings;
    QJsonObject m_currentSettings;
    bool m_settingsModified;
};

#endif // SETTINGS_DIALOG_H 