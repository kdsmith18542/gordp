#include "input_handler.h"
#include <QApplication>
#include <QWidget>
#include <QDebug>
#include <QMouseEvent>
#include <QKeyEvent>
#include <QWheelEvent>
#include <QTimer>
#include <QJsonObject>
#include <QJsonDocument>
#include <QSet>

InputHandler::InputHandler(QObject *parent)
    : QObject(parent)
    , m_inputCaptured(false)
    , m_mouseSensitivity(1.0)
    , m_keyboardRepeatDelay(500)
    , m_keyboardRepeatRate(30)
    , m_keyRepeatTimer(new QTimer(this))
    , m_lastPressedKey(0)
    , m_zoomLevel(1.0)
{
    // Setup key repeat timer
    connect(m_keyRepeatTimer, &QTimer::timeout, this, &InputHandler::onKeyRepeat);
    
    // Initialize resolutions
    m_localResolution = QSize(1024, 768);
    m_remoteResolution = QSize(1024, 768);
}

InputHandler::~InputHandler()
{
    releaseInput();
}

void InputHandler::handleMouseEvent(QMouseEvent *event)
{
    if (!m_inputCaptured) {
        return;
    }
    
    QPoint remotePos = convertToRemoteCoordinates(event->pos());
    
    switch (event->type()) {
    case QEvent::MouseButtonPress:
        m_pressedButtons |= event->button();
        sendMouseEvent(remotePos.x(), remotePos.y(), 
                      convertQtMouseButton(event->button()), true);
        break;
        
    case QEvent::MouseButtonRelease:
        m_pressedButtons &= ~event->button();
        sendMouseEvent(remotePos.x(), remotePos.y(), 
                      convertQtMouseButton(event->button()), false);
        break;
        
    case QEvent::MouseMove:
        if (m_pressedButtons != Qt::NoButton) {
            sendMouseEvent(remotePos.x(), remotePos.y(), 0, true);
        }
        break;
    }
    
    m_lastMousePos = event->pos();
}

void InputHandler::handleKeyEvent(QKeyEvent *event)
{
    if (!m_inputCaptured) {
        return;
    }
    
    int rdpKey = convertQtKeyToRDP(event->key());
    if (rdpKey == 0) {
        return; // Unsupported key
    }
    
    bool pressed = (event->type() == QEvent::KeyPress);
    
    if (pressed) {
        m_pressedKeys.insert(event->key());
        m_lastPressedKey = event->key();
        
        // Start key repeat for certain keys
        if (event->isAutoRepeat()) {
            m_keyRepeatTimer->start(m_keyboardRepeatRate);
        } else {
            m_keyRepeatTimer->start(m_keyboardRepeatDelay);
        }
    } else {
        m_pressedKeys.remove(event->key());
        m_keyRepeatTimer->stop();
    }
    
    sendKeyEvent(rdpKey, pressed);
}

void InputHandler::handleWheelEvent(QWheelEvent *event)
{
    if (!m_inputCaptured) {
        return;
    }
    
    int delta = event->angleDelta().y() / 120; // Convert to wheel clicks
    sendWheelEvent(delta);
}

void InputHandler::setMouseSensitivity(double sensitivity)
{
    m_mouseSensitivity = qBound(0.1, sensitivity, 5.0);
}

void InputHandler::setKeyboardRepeatDelay(int delay)
{
    m_keyboardRepeatDelay = qBound(100, delay, 2000);
}

void InputHandler::setKeyboardRepeatRate(int rate)
{
    m_keyboardRepeatRate = qBound(10, rate, 100);
}

void InputHandler::captureInput(bool capture)
{
    m_inputCaptured = capture;
    
    if (capture) {
        // Capture mouse and keyboard
        QApplication::setOverrideCursor(Qt::BlankCursor);
    } else {
        // Release capture
        releaseInput();
    }
}

void InputHandler::releaseInput()
{
    m_inputCaptured = false;
    m_pressedButtons = Qt::NoButton;
    m_pressedKeys.clear();
    m_keyRepeatTimer->stop();
    
    QApplication::restoreOverrideCursor();
}

void InputHandler::sendMouseEvent(int x, int y, int button, bool pressed)
{
    // Create RDP mouse event data
    QJsonObject eventData;
    eventData["type"] = "mouse";
    eventData["x"] = x;
    eventData["y"] = y;
    eventData["button"] = button;
    eventData["pressed"] = pressed;
    
    QJsonDocument doc(eventData);
    emit inputEventSent(doc.toJson());
    emit mouseEvent(x, y, button, pressed);
}

void InputHandler::sendKeyEvent(int key, bool pressed)
{
    // Create RDP key event data
    QJsonObject eventData;
    eventData["type"] = "keyboard";
    eventData["key"] = key;
    eventData["pressed"] = pressed;
    
    QJsonDocument doc(eventData);
    emit inputEventSent(doc.toJson());
    emit keyEvent(key, pressed);
}

void InputHandler::sendWheelEvent(int delta)
{
    // Create RDP wheel event data
    QJsonObject eventData;
    eventData["type"] = "wheel";
    eventData["delta"] = delta;
    
    QJsonDocument doc(eventData);
    emit inputEventSent(doc.toJson());
    emit wheelEvent(delta);
}

int InputHandler::convertQtKeyToRDP(int qtKey) const
{
    // Map Qt key codes to RDP virtual key codes
    switch (qtKey) {
    case Qt::Key_Escape: return 0x1B;
    case Qt::Key_Tab: return 0x09;
    case Qt::Key_CapsLock: return 0x14;
    case Qt::Key_Shift: return 0x10;
    case Qt::Key_Control: return 0x11;
    case Qt::Key_Alt: return 0x12;
    case Qt::Key_Backspace: return 0x08;
    case Qt::Key_Return: return 0x0D;
    case Qt::Key_Enter: return 0x0D;
    case Qt::Key_Space: return 0x20;
    case Qt::Key_Left: return 0x25;
    case Qt::Key_Up: return 0x26;
    case Qt::Key_Right: return 0x27;
    case Qt::Key_Down: return 0x28;
    case Qt::Key_Insert: return 0x2D;
    case Qt::Key_Delete: return 0x2E;
    case Qt::Key_Home: return 0x24;
    case Qt::Key_End: return 0x23;
    case Qt::Key_PageUp: return 0x21;
    case Qt::Key_PageDown: return 0x22;
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
        // For printable characters, use ASCII value
        if (qtKey >= Qt::Key_A && qtKey <= Qt::Key_Z) {
            return qtKey - Qt::Key_A + 'A';
        }
        if (qtKey >= Qt::Key_0 && qtKey <= Qt::Key_9) {
            return qtKey - Qt::Key_0 + '0';
        }
        return 0; // Unsupported key
    }
}

int InputHandler::convertQtMouseButton(Qt::MouseButton button) const
{
    switch (button) {
    case Qt::LeftButton: return 1;
    case Qt::RightButton: return 2;
    case Qt::MiddleButton: return 4;
    case Qt::XButton1: return 8;
    case Qt::XButton2: return 16;
    default: return 0;
    }
}

QPoint InputHandler::convertToRemoteCoordinates(const QPoint &localPoint) const
{
    if (m_remoteResolution.isEmpty() || m_localResolution.isEmpty()) {
        return localPoint;
    }
    
    double scaleX = static_cast<double>(m_remoteResolution.width()) / m_localResolution.width();
    double scaleY = static_cast<double>(m_remoteResolution.height()) / m_localResolution.height();
    
    int remoteX = static_cast<int>(localPoint.x() * scaleX * m_zoomLevel);
    int remoteY = static_cast<int>(localPoint.y() * scaleY * m_zoomLevel);
    
    return QPoint(remoteX, remoteY);
}

void InputHandler::onKeyRepeat()
{
    if (m_lastPressedKey != 0 && m_pressedKeys.contains(m_lastPressedKey)) {
        int rdpKey = convertQtKeyToRDP(m_lastPressedKey);
        if (rdpKey != 0) {
            sendKeyEvent(rdpKey, true);
        }
    }
}
