#ifndef RDP_DISPLAY_H
#define RDP_DISPLAY_H

#include <QWidget>
#include <QPixmap>
#include <QTimer>
#include <QMouseEvent>
#include <QKeyEvent>
#include <QWheelEvent>
#include <QPaintEvent>
#include <QResizeEvent>
#include <QFocusEvent>

class RDPDisplayWidget : public QWidget
{
    Q_OBJECT

public:
    explicit RDPDisplayWidget(QWidget *parent = nullptr);
    ~RDPDisplayWidget();

    // Display methods
    void updateBitmap(const QImage &image);
    void clearDisplay();
    void setZoomLevel(double zoom);
    void setFullscreen(bool fullscreen);
    
    // Getters
    double zoomLevel() const { return m_zoomLevel; }
    bool isFullscreen() const { return m_isFullscreen; }
    QSize remoteResolution() const { return m_remoteResolution; }

public slots:
    void handleResize(int width, int height);

signals:
    void mouseEvent(int x, int y, int button, bool pressed);
    void keyEvent(int key, bool pressed);
    void wheelEvent(int delta);
    void focusChanged(bool focused);

protected:
    // Event handlers
    void paintEvent(QPaintEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;
    void mousePressEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void wheelEvent(QWheelEvent *event) override;
    void keyPressEvent(QKeyEvent *event) override;
    void keyReleaseEvent(QKeyEvent *event) override;
    void focusInEvent(QFocusEvent *event) override;
    void focusOutEvent(QFocusEvent *event) override;

private:
    // Coordinate conversion
    QPoint convertToRemoteCoordinates(const QPoint &localPoint) const;
    QPoint convertFromRemoteCoordinates(const QPoint &remotePoint) const;
    
    // Mouse button conversion
    int convertQtMouseButton(Qt::MouseButton button) const;
    
    // Key conversion
    int convertQtKeyToRDP(int qtKey) const;
    
    // Display properties
    QPixmap m_displayPixmap;
    QTimer m_updateTimer;
    double m_zoomLevel;
    bool m_isFullscreen;
    QSize m_remoteResolution;
    QSize m_widgetSize;
    
    // Input state
    bool m_mouseCaptured;
    bool m_keyboardCaptured;
    Qt::MouseButtons m_pressedButtons;
    
    // Performance
    bool m_useHardwareAcceleration;
    QTimer m_performanceTimer;
};

#endif // RDP_DISPLAY_H 