#include "rdp_display.h"
#include <QPainter>
#include <QApplication>
#include <QScreen>
#include <QDebug>

RDPDisplayWidget::RDPDisplayWidget(QWidget *parent)
    : QWidget(parent)
    , m_zoomLevel(1.0)
    , m_isFullscreen(false)
    , m_remoteResolution(1024, 768)
    , m_widgetSize(800, 600)
    , m_mouseCaptured(false)
    , m_keyboardCaptured(false)
    , m_pressedButtons(Qt::NoButton)
    , m_useHardwareAcceleration(true)
{
    setFocusPolicy(Qt::StrongFocus);
    setMouseTracking(true);
    setMinimumSize(400, 300);
    
    // Setup update timer for smooth rendering
    connect(&m_updateTimer, &QTimer::timeout, this, [this]() {
        update();
    });
    m_updateTimer.setInterval(16); // ~60 FPS
    
    // Setup performance timer
    connect(&m_performanceTimer, &QTimer::timeout, this, [this]() {
        // Monitor performance metrics
    });
    m_performanceTimer.setInterval(1000);
    
    // Create a default display
    m_displayPixmap = QPixmap(m_remoteResolution);
    m_displayPixmap.fill(Qt::black);
}

RDPDisplayWidget::~RDPDisplayWidget()
{
}

void RDPDisplayWidget::updateBitmap(const QImage &image)
{
    if (image.isNull()) {
        qWarning() << "Received null image";
        return;
    }
    
    m_remoteResolution = image.size();
    m_displayPixmap = QPixmap::fromImage(image);
    
    // Trigger a repaint
    update();
}

void RDPDisplayWidget::clearDisplay()
{
    m_displayPixmap = QPixmap(m_remoteResolution);
    m_displayPixmap.fill(Qt::black);
    update();
}

void RDPDisplayWidget::setZoomLevel(double zoom)
{
    if (zoom < 0.1 || zoom > 5.0) {
        qWarning() << "Invalid zoom level:" << zoom;
        return;
    }
    
    m_zoomLevel = zoom;
    update();
}

void RDPDisplayWidget::setFullscreen(bool fullscreen)
{
    m_isFullscreen = fullscreen;
    
    if (fullscreen) {
        showFullScreen();
    } else {
        showNormal();
    }
}

void RDPDisplayWidget::handleResize(int width, int height)
{
    m_widgetSize = QSize(width, height);
    update();
}

void RDPDisplayWidget::paintEvent(QPaintEvent *event)
{
    Q_UNUSED(event)
    
    QPainter painter(this);
    painter.setRenderHint(QPainter::SmoothPixmapTransform, true);
    
    if (m_displayPixmap.isNull()) {
        // Draw placeholder
        painter.fillRect(rect(), Qt::black);
        painter.setPen(Qt::white);
        painter.drawText(rect(), Qt::AlignCenter, "No remote display");
        return;
    }
    
    // Calculate scaled size
    QSize scaledSize = m_remoteResolution * m_zoomLevel;
    
    // Center the image
    QRect targetRect = rect();
    if (scaledSize.width() < targetRect.width()) {
        targetRect.setX((targetRect.width() - scaledSize.width()) / 2);
        targetRect.setWidth(scaledSize.width());
    }
    if (scaledSize.height() < targetRect.height()) {
        targetRect.setY((targetRect.height() - scaledSize.height()) / 2);
        targetRect.setHeight(scaledSize.height());
    }
    
    // Draw the remote desktop
    painter.drawPixmap(targetRect, m_displayPixmap, m_displayPixmap.rect());
}

void RDPDisplayWidget::resizeEvent(QResizeEvent *event)
{
    QWidget::resizeEvent(event);
    m_widgetSize = event->size();
}

void RDPDisplayWidget::mousePressEvent(QMouseEvent *event)
{
    QPoint remotePoint = convertToRemoteCoordinates(event->pos());
    int button = convertQtMouseButton(event->button());
    
    emit mouseEvent(remotePoint.x(), remotePoint.y(), button, true);
    
    m_pressedButtons |= event->button();
    setFocus();
}

void RDPDisplayWidget::mouseReleaseEvent(QMouseEvent *event)
{
    QPoint remotePoint = convertToRemoteCoordinates(event->pos());
    int button = convertQtMouseButton(event->button());
    
    emit mouseEvent(remotePoint.x(), remotePoint.y(), button, false);
    
    m_pressedButtons &= ~event->button();
}

void RDPDisplayWidget::mouseMoveEvent(QMouseEvent *event)
{
    QPoint remotePoint = convertToRemoteCoordinates(event->pos());
    
    // Send mouse move event (button 0 for move)
    emit mouseEvent(remotePoint.x(), remotePoint.y(), 0, false);
}

void RDPDisplayWidget::wheelEvent(QWheelEvent *event)
{
    int delta = event->angleDelta().y();
    emit wheelEvent(delta);
}

void RDPDisplayWidget::keyPressEvent(QKeyEvent *event)
{
    int rdpKey = convertQtKeyToRDP(event->key());
    emit keyEvent(rdpKey, true);
}

void RDPDisplayWidget::keyReleaseEvent(QKeyEvent *event)
{
    int rdpKey = convertQtKeyToRDP(event->key());
    emit keyEvent(rdpKey, false);
}

void RDPDisplayWidget::focusInEvent(QFocusEvent *event)
{
    QWidget::focusInEvent(event);
    emit focusChanged(true);
}

void RDPDisplayWidget::focusOutEvent(QFocusEvent *event)
{
    QWidget::focusOutEvent(event);
    emit focusChanged(false);
}

QPoint RDPDisplayWidget::convertToRemoteCoordinates(const QPoint &localPoint) const
{
    if (m_remoteResolution.isEmpty()) {
        return localPoint;
    }
    
    // Calculate the display area
    QSize scaledSize = m_remoteResolution * m_zoomLevel;
    QRect displayRect = rect();
    
    if (scaledSize.width() < displayRect.width()) {
        displayRect.setX((displayRect.width() - scaledSize.width()) / 2);
        displayRect.setWidth(scaledSize.width());
    }
    if (scaledSize.height() < displayRect.height()) {
        displayRect.setY((displayRect.height() - scaledSize.height()) / 2);
        displayRect.setHeight(scaledSize.height());
    }
    
    // Convert coordinates
    QPoint relativePoint = localPoint - displayRect.topLeft();
    QPoint remotePoint(
        relativePoint.x() / m_zoomLevel,
        relativePoint.y() / m_zoomLevel
    );
    
    // Clamp to remote resolution
    remotePoint.setX(qBound(0, remotePoint.x(), m_remoteResolution.width() - 1));
    remotePoint.setY(qBound(0, remotePoint.y(), m_remoteResolution.height() - 1));
    
    return remotePoint;
}

QPoint RDPDisplayWidget::convertFromRemoteCoordinates(const QPoint &remotePoint) const
{
    if (m_remoteResolution.isEmpty()) {
        return remotePoint;
    }
    
    // Calculate the display area
    QSize scaledSize = m_remoteResolution * m_zoomLevel;
    QRect displayRect = rect();
    
    if (scaledSize.width() < displayRect.width()) {
        displayRect.setX((displayRect.width() - scaledSize.width()) / 2);
        displayRect.setWidth(scaledSize.width());
    }
    if (scaledSize.height() < displayRect.height()) {
        displayRect.setY((displayRect.height() - scaledSize.height()) / 2);
        displayRect.setHeight(scaledSize.height());
    }
    
    // Convert coordinates
    QPoint localPoint(
        remotePoint.x() * m_zoomLevel + displayRect.x(),
        remotePoint.y() * m_zoomLevel + displayRect.y()
    );
    
    return localPoint;
}

int RDPDisplayWidget::convertQtMouseButton(Qt::MouseButton button) const
{
    switch (button) {
        case Qt::LeftButton: return 1;
        case Qt::RightButton: return 2;
        case Qt::MiddleButton: return 3;
        case Qt::XButton1: return 4;
        case Qt::XButton2: return 5;
        default: return 0;
    }
}

int RDPDisplayWidget::convertQtKeyToRDP(int qtKey) const
{
    // This is a simplified key mapping
    // In a real implementation, you would have a comprehensive mapping
    // from Qt key codes to RDP virtual key codes
    
    switch (qtKey) {
        case Qt::Key_Escape: return 0x1B;
        case Qt::Key_Return: return 0x0D;
        case Qt::Key_Tab: return 0x09;
        case Qt::Key_Backspace: return 0x08;
        case Qt::Key_Delete: return 0x2E;
        case Qt::Key_Insert: return 0x2D;
        case Qt::Key_Home: return 0x24;
        case Qt::Key_End: return 0x23;
        case Qt::Key_PageUp: return 0x21;
        case Qt::Key_PageDown: return 0x22;
        case Qt::Key_Up: return 0x26;
        case Qt::Key_Down: return 0x28;
        case Qt::Key_Left: return 0x25;
        case Qt::Key_Right: return 0x27;
        case Qt::Key_F1: return 0x70;
        case Qt::Key_F2: return 0x71;
        case Qt::Key_F3: return 0x72;
        case Qt::Key_F4: return 0x73;
        case Qt::Key_F5: return 0x74;
        case Qt::Key_F6: return 0x75;
        case Qt::Key_F7: return 0x76;
        case Qt::Key_F8: return 0x77;
        case Qt::Key_F9: return 0x78;
        case Qt::Key_F10: return 0x79;
        case Qt::Key_F11: return 0x7A;
        case Qt::Key_F12: return 0x7B;
        default:
            // For printable characters, return the ASCII value
            if (qtKey >= 0x20 && qtKey <= 0x7E) {
                return qtKey;
            }
            return 0;
    }
} 