#ifndef INPUT_HANDLER_H
#define INPUT_HANDLER_H

#include <QObject>
#include <QMouseEvent>
#include <QKeyEvent>
#include <QWheelEvent>
#include <QTimer>
#include <QPoint>
#include <QSet>
#include <QSize>

class InputHandler : public QObject
{
    Q_OBJECT

public:
    explicit InputHandler(QObject *parent = nullptr);
    ~InputHandler();

    // Input handling methods
    void handleMouseEvent(QMouseEvent *event);
    void handleKeyEvent(QKeyEvent *event);
    void handleWheelEvent(QWheelEvent *event);
    
    // Configuration
    void setMouseSensitivity(double sensitivity);
    void setKeyboardRepeatDelay(int delay);
    void setKeyboardRepeatRate(int rate);
    
    // State management
    void captureInput(bool capture);
    void releaseInput();
    bool isInputCaptured() const { return m_inputCaptured; }

signals:
    void inputEventSent(const QByteArray &eventData);
    void mouseEvent(int x, int y, int button, bool pressed);
    void keyEvent(int key, bool pressed);
    void wheelEvent(int delta);

private slots:
    void onKeyRepeat();

private:
    // Input event conversion
    void sendMouseEvent(int x, int y, int button, bool pressed);
    void sendKeyEvent(int key, bool pressed);
    void sendWheelEvent(int delta);
    
    // Key mapping
    int convertQtKeyToRDP(int qtKey) const;
    int convertQtMouseButton(Qt::MouseButton button) const;
    
    // Coordinate conversion
    QPoint convertToRemoteCoordinates(const QPoint &localPoint) const;
    
    // Input state
    bool m_inputCaptured;
    Qt::MouseButtons m_pressedButtons;
    QSet<int> m_pressedKeys;
    
    // Configuration
    double m_mouseSensitivity;
    int m_keyboardRepeatDelay;
    int m_keyboardRepeatRate;
    
    // Key repeat handling
    QTimer *m_keyRepeatTimer;
    int m_lastPressedKey;
    QPoint m_lastMousePos;
    
    // Remote desktop properties
    QSize m_remoteResolution;
    QSize m_localResolution;
    double m_zoomLevel;
};

#endif // INPUT_HANDLER_H
