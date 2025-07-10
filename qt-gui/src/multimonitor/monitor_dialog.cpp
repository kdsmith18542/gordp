#include "monitor_dialog.h"
#include "ui_monitor_dialog.h"
#include "monitor_layout_preview_widget.h"
#include <QMessageBox>
#include <QJsonDocument>
#include <QFileDialog>
#include <QStandardPaths>
#include <QScreen>
#include <QGuiApplication>
#include <QDebug>

MonitorDialog::MonitorDialog(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::MonitorDialog)
    , m_settings(new QSettings("GoRDP", "MultiMonitor"))
    , m_selectedMonitors()
    , m_monitorLayout()
{
    ui->setupUi(this);
    setupConnections();
    detectMonitors();
    loadSettings();
}

MonitorDialog::~MonitorDialog()
{
    delete ui;
    delete m_settings;
}

void MonitorDialog::detectMonitors()
{
    m_monitors = QJsonArray();
    
    // Get available screens
    QList<QScreen*> screens = QGuiApplication::screens();
    
    for (QScreen *screen : screens) {
        QJsonObject monitor;
        monitor["name"] = screen->name();
        monitor["geometry"] = QJsonObject{
            {"x", screen->geometry().x()},
            {"y", screen->geometry().y()},
            {"width", screen->geometry().width()},
            {"height", screen->geometry().height()}
        };
        monitor["available"] = true;
        monitor["primary"] = (screen == QGuiApplication::primaryScreen());
        
        m_monitors.append(monitor);
    }
    
    updateLayoutPreview();
    updateStatus(QString("Detected %1 monitor(s)").arg(m_monitors.size()));
}

void MonitorDialog::loadSettings()
{
    // Load selected monitors
    QJsonArray selectedMonitors = QJsonDocument::fromJson(
        m_settings->value("selectedMonitors", "[]").toByteArray()).array();
    
    for (int i = 0; i < ui->monitorList->count(); ++i) {
        QListWidgetItem *item = ui->monitorList->item(i);
        int monitorIndex = item->data(Qt::UserRole).toInt();
        
        bool selected = false;
        for (const QJsonValue &value : selectedMonitors) {
            if (value.toInt() == monitorIndex) {
                selected = true;
                break;
            }
        }
        
        item->setCheckState(selected ? Qt::Checked : Qt::Unchecked);
    }
    
    // Load monitor layout
    QJsonObject layout = QJsonDocument::fromJson(
        m_settings->value("monitorLayout", "{}").toByteArray()).object();
    
    if (!layout.isEmpty()) {
        m_monitorLayout = layout;
    }
    
    updateLayoutPreview();
}

void MonitorDialog::saveSettings()
{
    // Save selected monitors
    QJsonArray selectedMonitors;
    for (int i = 0; i < ui->monitorList->count(); ++i) {
        QListWidgetItem *item = ui->monitorList->item(i);
        if (item->checkState() == Qt::Checked) {
            int monitorIndex = item->data(Qt::UserRole).toInt();
            selectedMonitors.append(monitorIndex);
        }
    }
    
    m_settings->setValue("selectedMonitors", QJsonDocument(selectedMonitors).toJson());
    m_settings->setValue("monitorLayout", QJsonDocument(m_monitorLayout).toJson());
    m_settings->sync();
}

QJsonObject MonitorDialog::getMonitorConfiguration() const
{
    QJsonObject config;
    
    // Get selected monitors
    QJsonArray selectedMonitors;
    for (int i = 0; i < ui->monitorList->count(); ++i) {
        QListWidgetItem *item = ui->monitorList->item(i);
        if (item->checkState() == Qt::Checked) {
            int monitorIndex = item->data(Qt::UserRole).toInt();
            selectedMonitors.append(m_monitors[monitorIndex]);
        }
    }
    
    config["selectedMonitors"] = selectedMonitors;
    config["layout"] = m_monitorLayout;
    
    return config;
}

void MonitorDialog::onDetectMonitorsClicked()
{
    detectMonitors();
    QMessageBox::information(this, "Monitors Detected", 
                           QString("Found %1 monitor(s)").arg(m_monitors.size()));
}

void MonitorDialog::onApplyLayoutClicked()
{
    // Validate selection
    int selectedCount = 0;
    for (int i = 0; i < ui->monitorList->count(); ++i) {
        if (ui->monitorList->item(i)->checkState() == Qt::Checked) {
            selectedCount++;
        }
    }
    
    if (selectedCount == 0) {
        QMessageBox::warning(this, "No Monitors Selected", 
                           "Please select at least one monitor for the RDP session.");
        return;
    }
    
    saveSettings();
    
    // Apply configuration to active connection
    QJsonObject config = getMonitorConfiguration();
    emit configurationApplied(config);
    
    QMessageBox::information(this, "Layout Applied", 
                           "Monitor configuration has been applied to the current connection.");
}

void MonitorDialog::onResetLayoutClicked()
{
    int result = QMessageBox::question(this, "Reset Layout", 
                                      "Are you sure you want to reset the monitor layout to defaults?",
                                      QMessageBox::Yes | QMessageBox::No);
    
    if (result == QMessageBox::Yes) {
        // Reset to primary monitor only
        for (int i = 0; i < ui->monitorList->count(); ++i) {
            QListWidgetItem *item = ui->monitorList->item(i);
            int monitorIndex = item->data(Qt::UserRole).toInt();
            
            // Check if this is the primary monitor
            if (monitorIndex == 0) { // Assuming primary is index 0
                item->setCheckState(Qt::Checked);
            } else {
                item->setCheckState(Qt::Unchecked);
            }
        }
        
        m_monitorLayout = QJsonObject();
        updateLayoutPreview();
    }
}

void MonitorDialog::onCloseClicked()
{
    accept();
}

void MonitorDialog::onItemSelectionChanged()
{
    updateLayoutPreview();
}

void MonitorDialog::onMonitorSelectionChanged() {
    // Update selected monitors based on UI selection
    m_selectedMonitors = QJsonArray();
    QList<QListWidgetItem*> selectedItems = ui->monitorList->selectedItems();
    for (QListWidgetItem* item : selectedItems) {
        int row = ui->monitorList->row(item);
        if (row >= 0 && row < m_monitors.size()) {
            m_selectedMonitors.append(m_monitors[row]);
        }
    }
    updateLayoutPreview();
    updateStatus(QString("%1 monitor(s) selected").arg(m_selectedMonitors.size()));
}

void MonitorDialog::onSelectAllClicked() {
    ui->monitorList->selectAll();
    onMonitorSelectionChanged();
}

void MonitorDialog::onClearSelectionClicked() {
    ui->monitorList->clearSelection();
    onMonitorSelectionChanged();
}

void MonitorDialog::onApplyClicked() {
    // Save selected monitors and emit signal
    saveSettings();
    emit monitorsSelected(m_selectedMonitors);
    accept();
}

void MonitorDialog::onCancelClicked() {
    reject();
}

void MonitorDialog::updateStatus(const QString& statusText) {
    if (ui->statusLabel) {
        ui->statusLabel->setText(statusText);
    }
}

void MonitorDialog::setupConnections()
{
    connect(ui->detectMonitorsButton, &QPushButton::clicked, 
            this, &MonitorDialog::onDetectMonitorsClicked);
    connect(ui->applyLayoutButton, &QPushButton::clicked, 
            this, &MonitorDialog::onApplyLayoutClicked);
    connect(ui->resetLayoutButton, &QPushButton::clicked, 
            this, &MonitorDialog::onResetLayoutClicked);
    connect(ui->closeButton, &QPushButton::clicked, 
            this, &MonitorDialog::onCloseClicked);
    
    connect(ui->monitorList, &QListWidget::itemSelectionChanged, 
            this, &MonitorDialog::onItemSelectionChanged);
}

void MonitorDialog::updateLayoutPreview()
{
    // Create real visual monitor layout preview
    if (!m_layoutPreviewWidget) {
        m_layoutPreviewWidget = new MonitorLayoutPreviewWidget(this);
        
        // Replace placeholder with real preview widget
        QVBoxLayout *layout = qobject_cast<QVBoxLayout*>(ui->layoutPlaceholder->parentWidget()->layout());
        if (layout) {
            int index = layout->indexOf(ui->layoutPlaceholder);
            if (index != -1) {
                layout->removeWidget(ui->layoutPlaceholder);
                ui->layoutPlaceholder->hide();
                layout->insertWidget(index, m_layoutPreviewWidget);
            }
        }
    }
    
    // Get selected monitors
    QJsonArray selectedMonitors;
    for (int i = 0; i < ui->monitorList->count(); ++i) {
        QListWidgetItem *item = ui->monitorList->item(i);
        if (item->checkState() == Qt::Checked) {
            int monitorIndex = item->data(Qt::UserRole).toInt();
            selectedMonitors.append(m_monitors[monitorIndex]);
        }
    }
    
    // Update preview widget with selected monitors
    m_layoutPreviewWidget->setMonitors(selectedMonitors);
    
    // Update status text
    QString statusText;
    if (selectedMonitors.isEmpty()) {
        statusText = "No monitors selected";
    } else if (selectedMonitors.size() == 1) {
        QJsonObject monitor = selectedMonitors[0].toObject();
        QJsonObject resolution = monitor["resolution"].toObject();
        statusText = QString("Single monitor mode (%1x%2)")
            .arg(resolution["width"].toInt())
            .arg(resolution["height"].toInt());
    } else {
        statusText = QString("Multi-monitor mode (%1 monitors)").arg(selectedMonitors.size());
        
        // Calculate total resolution
        int totalWidth = 0;
        int maxHeight = 0;
        
        for (const QJsonValue &monitor : selectedMonitors) {
            QJsonObject monitorObj = monitor.toObject();
            QJsonObject resolution = monitorObj["resolution"].toObject();
            totalWidth += resolution["width"].toInt();
            maxHeight = qMax(maxHeight, resolution["height"].toInt());
        }
        
        statusText += QString(" - Total: %1x%2").arg(totalWidth).arg(maxHeight);
    }
    
    // Update status label
    if (ui->statusLabel) {
        ui->statusLabel->setText(statusText);
    }
}

void MonitorDialog::mousePressEvent(QMouseEvent *event)
{
    // Handle mouse events for interactive layout editing
    // This would allow users to drag monitors around in the layout preview
    QDialog::mousePressEvent(event);
}

void MonitorDialog::mouseMoveEvent(QMouseEvent *event)
{
    // Handle mouse movement for interactive layout editing
    QDialog::mouseMoveEvent(event);
}

void MonitorDialog::mouseReleaseEvent(QMouseEvent *event)
{
    // Handle mouse release for interactive layout editing
    QDialog::mouseReleaseEvent(event);
}
