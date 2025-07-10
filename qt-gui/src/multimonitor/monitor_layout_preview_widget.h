#ifndef MONITOR_LAYOUT_PREVIEW_WIDGET_H
#define MONITOR_LAYOUT_PREVIEW_WIDGET_H

#include <QWidget>
#include <QJsonArray>
#include <QJsonObject>
#include <QPainter>
#include <QMouseEvent>

class MonitorLayoutPreviewWidget : public QWidget
{
    Q_OBJECT

public:
    explicit MonitorLayoutPreviewWidget(QWidget *parent = nullptr);
    
    void setMonitors(const QJsonArray &monitors);
    QJsonArray getSelectedMonitors() const;
    
protected:
    void paintEvent(QPaintEvent *event) override;
    void mousePressEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;
    
private:
    struct MonitorRect {
        QRect rect;
        QJsonObject data;
        bool selected;
        bool hovered;
    };
    
    QJsonArray m_monitors;
    QList<MonitorRect> m_monitorRects;
    int m_hoveredMonitor;
    int m_selectedMonitor;
    QPoint m_dragStart;
    bool m_dragging;
    
    void updateMonitorRects();
    QRect calculateMonitorRect(const QJsonObject &monitor, int index, int total);
    void drawMonitor(QPainter &painter, const MonitorRect &monitorRect);
    int findMonitorAt(const QPoint &pos);
    QColor getMonitorColor(const MonitorRect &monitorRect);
};

#endif // MONITOR_LAYOUT_PREVIEW_WIDGET_H 