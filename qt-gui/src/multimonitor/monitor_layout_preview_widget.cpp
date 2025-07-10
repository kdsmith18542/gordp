#include "monitor_layout_preview_widget.h"
#include <QPainter>
#include <QMouseEvent>
#include <QApplication>
#include <QScreen>
#include <QDebug>

MonitorLayoutPreviewWidget::MonitorLayoutPreviewWidget(QWidget *parent)
    : QWidget(parent)
    , m_hoveredMonitor(-1)
    , m_selectedMonitor(-1)
    , m_dragging(false)
{
    setMinimumSize(300, 200);
    setMouseTracking(true);
    setFocusPolicy(Qt::StrongFocus);
}

void MonitorLayoutPreviewWidget::setMonitors(const QJsonArray &monitors)
{
    m_monitors = monitors;
    updateMonitorRects();
    update();
}

QJsonArray MonitorLayoutPreviewWidget::getSelectedMonitors() const
{
    QJsonArray selected;
    for (const MonitorRect &rect : m_monitorRects) {
        if (rect.selected) {
            selected.append(rect.data);
        }
    }
    return selected;
}

void MonitorLayoutPreviewWidget::paintEvent(QPaintEvent *event)
{
    Q_UNUSED(event)
    
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing);
    
    // Draw background
    painter.fillRect(rect(), QColor(40, 40, 40));
    
    // Draw grid
    painter.setPen(QPen(QColor(60, 60, 60), 1));
    for (int x = 0; x < width(); x += 20) {
        painter.drawLine(x, 0, x, height());
    }
    for (int y = 0; y < height(); y += 20) {
        painter.drawLine(0, y, width(), y);
    }
    
    // Draw monitors
    for (const MonitorRect &monitorRect : m_monitorRects) {
        drawMonitor(painter, monitorRect);
    }
    
    // Draw instructions if no monitors
    if (m_monitorRects.isEmpty()) {
        painter.setPen(QColor(150, 150, 150));
        painter.setFont(QFont("Arial", 12));
        painter.drawText(rect(), Qt::AlignCenter, "No monitors available");
    }
}

void MonitorLayoutPreviewWidget::mousePressEvent(QMouseEvent *event)
{
    if (event->button() == Qt::LeftButton) {
        int monitorIndex = findMonitorAt(event->pos());
        if (monitorIndex != -1) {
            m_selectedMonitor = monitorIndex;
            m_dragging = true;
            m_dragStart = event->pos();
            
            // Toggle selection
            m_monitorRects[monitorIndex].selected = !m_monitorRects[monitorIndex].selected;
            update();
        }
    }
}

void MonitorLayoutPreviewWidget::mouseMoveEvent(QMouseEvent *event)
{
    int monitorIndex = findMonitorAt(event->pos());
    
    // Update hover state
    for (int i = 0; i < m_monitorRects.size(); ++i) {
        m_monitorRects[i].hovered = (i == monitorIndex);
    }
    
    m_hoveredMonitor = monitorIndex;
    update();
}

void MonitorLayoutPreviewWidget::mouseReleaseEvent(QMouseEvent *event)
{
    Q_UNUSED(event)
    m_dragging = false;
}

void MonitorLayoutPreviewWidget::updateMonitorRects()
{
    m_monitorRects.clear();
    
    for (int i = 0; i < m_monitors.size(); ++i) {
        QJsonObject monitor = m_monitors[i].toObject();
        MonitorRect rect;
        rect.data = monitor;
        rect.selected = false;
        rect.hovered = false;
        rect.rect = calculateMonitorRect(monitor, i, m_monitors.size());
        m_monitorRects.append(rect);
    }
}

QRect MonitorLayoutPreviewWidget::calculateMonitorRect(const QJsonObject &monitor, int index, int total)
{
    QJsonObject resolution = monitor["resolution"].toObject();
    int width = resolution["width"].toInt();
    int height = resolution["height"].toInt();
    
    // Scale down for preview (maintain aspect ratio)
    int maxPreviewWidth = this->width() - 40;
    int maxPreviewHeight = this->height() - 40;
    
    double scaleX = static_cast<double>(maxPreviewWidth) / width;
    double scaleY = static_cast<double>(maxPreviewHeight) / height;
    double scale = qMin(scaleX, scaleY);
    
    int previewWidth = static_cast<int>(width * scale);
    int previewHeight = static_cast<int>(height * scale);
    
    // Position monitors in a grid layout
    int cols = qMin(total, 3);
    int rows = (total + cols - 1) / cols;
    
    int col = index % cols;
    int row = index / cols;
    
    int x = 20 + col * (previewWidth + 20);
    int y = 20 + row * (previewHeight + 20);
    
    return QRect(x, y, previewWidth, previewHeight);
}

void MonitorLayoutPreviewWidget::drawMonitor(QPainter &painter, const MonitorRect &monitorRect)
{
    QRect rect = monitorRect.rect;
    QColor color = getMonitorColor(monitorRect);
    
    // Draw monitor border
    painter.setPen(QPen(color, 2));
    painter.setBrush(QBrush(color.lighter(120)));
    painter.drawRect(rect);
    
    // Draw monitor info
    QJsonObject resolution = monitorRect.data["resolution"].toObject();
    QString info = QString("%1x%2")
        .arg(resolution["width"].toInt())
        .arg(resolution["height"].toInt());
    
    painter.setPen(Qt::white);
    painter.setFont(QFont("Arial", 8));
    painter.drawText(rect, Qt::AlignCenter, info);
    
    // Draw selection indicator
    if (monitorRect.selected) {
        painter.setPen(QPen(Qt::green, 3));
        painter.setBrush(Qt::NoBrush);
        painter.drawRect(rect.adjusted(-2, -2, 2, 2));
    }
    
    // Draw hover effect
    if (monitorRect.hovered) {
        painter.setPen(QPen(Qt::yellow, 2));
        painter.setBrush(Qt::NoBrush);
        painter.drawRect(rect.adjusted(-1, -1, 1, 1));
    }
}

int MonitorLayoutPreviewWidget::findMonitorAt(const QPoint &pos)
{
    for (int i = 0; i < m_monitorRects.size(); ++i) {
        if (m_monitorRects[i].rect.contains(pos)) {
            return i;
        }
    }
    return -1;
}

QColor MonitorLayoutPreviewWidget::getMonitorColor(const MonitorRect &monitorRect)
{
    if (monitorRect.selected) {
        return QColor(0, 150, 0); // Green for selected
    } else if (monitorRect.hovered) {
        return QColor(150, 150, 0); // Yellow for hovered
    } else {
        return QColor(100, 100, 100); // Gray for normal
    }
} 