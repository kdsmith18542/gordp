#ifndef MONITOR_DIALOG_H
#define MONITOR_DIALOG_H

#include <QDialog>
#include <QJsonArray>
#include <QJsonObject>
#include <QSettings>
#include <QListWidget>
#include <QLabel>
#include <QVBoxLayout>
#include <QMouseEvent>

namespace Ui {
class MonitorDialog;
}

class MonitorLayoutPreviewWidget;

class MonitorDialog : public QDialog
{
    Q_OBJECT

public:
    explicit MonitorDialog(QWidget *parent = nullptr);
    ~MonitorDialog();

    QJsonArray getSelectedMonitors() const;
    void setMonitors(const QJsonArray &monitors);
    QJsonObject getMonitorConfiguration() const;

signals:
    void monitorsSelected(const QJsonArray &monitors);
    void configurationApplied(const QJsonObject &configuration);

private slots:
    void onMonitorSelectionChanged();
    void onSelectAllClicked();
    void onClearSelectionClicked();
    void onApplyClicked();
    void onCancelClicked();
    void onDetectMonitorsClicked();
    void onApplyLayoutClicked();
    void onResetLayoutClicked();
    void onCloseClicked();
    void onItemSelectionChanged();

private:
    void setupConnections();
    void loadMonitors();
    void updateLayoutPreview();
    void saveSettings();
    void loadSettings();
    void detectMonitors();
    void updateStatus(const QString &statusText);
    
    // Mouse event handling
    void mousePressEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;

    Ui::MonitorDialog *ui;
    QJsonArray m_monitors;
    QJsonArray m_selectedMonitors;
    QJsonObject m_monitorLayout;
    MonitorLayoutPreviewWidget *m_layoutPreviewWidget;
    QSettings *m_settings;
};

#endif // MONITOR_DIALOG_H 