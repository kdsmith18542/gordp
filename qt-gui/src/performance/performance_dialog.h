#ifndef PERFORMANCE_DIALOG_H
#define PERFORMANCE_DIALOG_H

#include <QDialog>
#include <QTimer>
#include <QJsonObject>
#include <QJsonArray>
#include <QList>
#include <QStandardPaths>
#include <QProcess>
#include <QChartView>
#include <QChart>
#include <QLineSeries>
#include <QValueAxis>
#include <QCoreApplication>

QT_BEGIN_NAMESPACE
class QChartView;
class QChart;
class QLineSeries;
class QValueAxis;
QT_END_NAMESPACE

namespace Ui {
class PerformanceDialog;
}

class GoRDPBridge;

class PerformanceDialog : public QDialog
{
    Q_OBJECT

public:
    explicit PerformanceDialog(QWidget *parent = nullptr);
    ~PerformanceDialog();

    // Performance monitoring
    void startMonitoring();
    void stopMonitoring();
    void updateStats(const QJsonObject &stats);

signals:
    void refreshRequested();
    void exportRequested();

private slots:
    void onRefreshClicked();
    void onExportClicked();
    void onCloseClicked();
    void onUpdateTimer();

private:
    void setupConnections();
    void updateUI();
    void exportData();
    void setupGraphs();
    void updateStats();
    void addToHistory(const QJsonObject &stats);
    void updateGraphs();
    bool exportData(const QString &fileName);
    QString formatBandwidth(double bandwidth);
    void paintEvent(QPaintEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;
    
    Ui::PerformanceDialog *ui;
    QTimer *m_updateTimer;
    QJsonObject m_currentStats;
    QList<QJsonObject> m_historyData;
    int m_maxHistoryPoints;
    bool m_monitoring;
    
    // Qt Charts members
    QChartView *m_chartView;
    QChart *m_chart;
    QLineSeries *m_bandwidthSeries;
    QLineSeries *m_latencySeries;
    QLineSeries *m_fpsSeries;
    QLineSeries *m_cpuSeries;
    QLineSeries *m_memorySeries;
    QValueAxis *m_axisX;
    QValueAxis *m_axisY;
    int m_timeCounter;
    int m_maxDataPoints;
    GoRDPBridge *m_gordpBridge;
};

#endif // PERFORMANCE_DIALOG_H
