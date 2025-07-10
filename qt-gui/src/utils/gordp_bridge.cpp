#include "gordp_bridge.h"
#include <QStandardPaths>
#include <QDir>
#include <QDebug>
#include <QJsonParseError>
#include <QBuffer>
#include <QIODevice>

GoRDPBridge::GoRDPBridge(QObject *parent)
    : QObject(parent)
    , m_networkManager(new QNetworkAccessManager(this))
    , m_websocket(new QWebSocket(QString(), QWebSocketProtocol::VersionLatest, this))
    , m_gordpProcess(new QProcess(this))
    , m_apiPort(8080)
    , m_isConnected(false)
    , m_apiRunning(false)
    , m_performanceTimer(new QTimer(this))
{
    m_apiUrl = QString("http://localhost:%1").arg(m_apiPort);
    
    // Connect signals
    connect(m_networkManager, &QNetworkAccessManager::finished,
            this, &GoRDPBridge::onNetworkReplyFinished);
    
    connect(m_websocket, &QWebSocket::connected,
            this, &GoRDPBridge::onWebSocketConnected);
    connect(m_websocket, &QWebSocket::disconnected,
            this, &GoRDPBridge::onWebSocketDisconnected);
    connect(m_websocket, &QWebSocket::textMessageReceived,
            this, &GoRDPBridge::onWebSocketTextMessageReceived);
    connect(m_websocket, &QWebSocket::binaryMessageReceived,
            this, &GoRDPBridge::onWebSocketBinaryMessageReceived);
    
    connect(m_gordpProcess, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished),
            this, &GoRDPBridge::onProcessFinished);
    connect(m_gordpProcess, &QProcess::errorOccurred,
            this, &GoRDPBridge::onProcessError);
    
    // Setup performance timer
    connect(m_performanceTimer, &QTimer::timeout,
            this, &GoRDPBridge::getPerformanceStats);
    m_performanceTimer->setInterval(1000); // Update every second
}

GoRDPBridge::~GoRDPBridge()
{
    stopGoRDPAPI();
}

bool GoRDPBridge::checkGoRDPAvailability()
{
    QString executablePath = getGoRDPExecutablePath();
    if (executablePath.isEmpty()) {
        qWarning() << "GoRDP executable not found in PATH";
        return false;
    }
    
    qDebug() << "GoRDP executable found at:" << executablePath;
    return true;
}

void GoRDPBridge::startGoRDPAPI()
{
    if (m_apiRunning) {
        qDebug() << "GoRDP API already running";
        return;
    }
    
    QString executablePath = getGoRDPExecutablePath();
    if (executablePath.isEmpty()) {
        emit apiError("GoRDP executable not found");
        return;
    }
    
    // Start GoRDP API server
    QStringList arguments;
    arguments << "--api" << "--port" << QString::number(m_apiPort);
    
    m_gordpProcess->start(executablePath, arguments);
    
    if (m_gordpProcess->waitForStarted()) {
        m_apiRunning = true;
        qDebug() << "GoRDP API server started";
        
        // Setup WebSocket connection for real-time updates
        setupWebSocket();
        
        emit apiStarted();
    } else {
        emit apiError("Failed to start GoRDP API server");
    }
}

void GoRDPBridge::stopGoRDPAPI()
{
    if (m_gordpProcess->state() != QProcess::NotRunning) {
        m_gordpProcess->terminate();
        if (!m_gordpProcess->waitForFinished(5000)) {
            m_gordpProcess->kill();
        }
    }
    
    m_websocket->close();
    m_apiRunning = false;
    m_isConnected = false;
    
    emit apiStopped();
}

void GoRDPBridge::connectToServer(const QString &server, int port, 
                                 const QString &username, const QString &password,
                                 const QJsonObject &options)
{
    if (!m_apiRunning) {
        emit connectionError("GoRDP API not running");
        return;
    }
    
    m_currentServer = server;
    m_currentPort = port;
    m_currentUsername = username;
    
    QJsonObject requestData;
    requestData["server"] = server;
    requestData["port"] = port;
    requestData["username"] = username;
    requestData["password"] = password;
    
    if (!options.isEmpty()) {
        requestData["options"] = options;
    }
    
    sendHttpRequest("/api/connect", requestData);
}

void GoRDPBridge::disconnectFromServer()
{
    if (m_isConnected) {
        sendHttpRequest("/api/disconnect");
        m_isConnected = false;
        emit connectionStatusChanged(false);
    }
}

void GoRDPBridge::sendMouseEvent(int x, int y, int button, bool pressed)
{
    if (m_websocket->state() == QAbstractSocket::ConnectedState) {
        QJsonObject event;
        event["type"] = "mouse";
        event["x"] = x;
        event["y"] = y;
        event["button"] = button;
        event["pressed"] = pressed;
        
        m_websocket->sendTextMessage(QJsonDocument(event).toJson());
    }
}

void GoRDPBridge::sendKeyEvent(int key, bool pressed)
{
    if (m_websocket->state() == QAbstractSocket::ConnectedState) {
        QJsonObject event;
        event["type"] = "keyboard";
        event["key"] = key;
        event["pressed"] = pressed;
        
        m_websocket->sendTextMessage(QJsonDocument(event).toJson());
    }
}

void GoRDPBridge::sendWheelEvent(int delta)
{
    if (m_websocket->state() == QAbstractSocket::ConnectedState) {
        QJsonObject event;
        event["type"] = "wheel";
        event["delta"] = delta;
        
        m_websocket->sendTextMessage(QJsonDocument(event).toJson());
    }
}

void GoRDPBridge::saveConnectionHistory(const QJsonObject &connection)
{
    QJsonObject requestData;
    requestData["connection"] = connection;
    sendHttpRequest("/api/history/save", requestData);
}

void GoRDPBridge::loadConnectionHistory()
{
    sendHttpRequest("/api/history/load");
}

void GoRDPBridge::saveFavorites(const QJsonArray &favorites)
{
    QJsonObject requestData;
    requestData["favorites"] = favorites;
    sendHttpRequest("/api/favorites/save", requestData);
}

void GoRDPBridge::loadFavorites()
{
    sendHttpRequest("/api/favorites/load");
}

void GoRDPBridge::getPerformanceStats()
{
    if (m_isConnected) {
        sendHttpRequest("/api/performance/stats");
    }
}

void GoRDPBridge::getConnectionStatus()
{
    sendHttpRequest("/api/status");
}

void GoRDPBridge::onNetworkReplyFinished(QNetworkReply *reply)
{
    reply->deleteLater();
    
    if (reply->error() != QNetworkReply::NoError) {
        emit connectionError(reply->errorString());
        return;
    }
    
    QByteArray data = reply->readAll();
    QJsonParseError parseError;
    QJsonDocument response = QJsonDocument::fromJson(data, &parseError);
    
    if (parseError.error != QJsonParseError::NoError) {
        emit connectionError("Invalid JSON response");
        return;
    }
    
    handleHttpResponse(response);
}

void GoRDPBridge::onWebSocketConnected()
{
    qDebug() << "WebSocket connected to GoRDP API";
}

void GoRDPBridge::onWebSocketDisconnected()
{
    qDebug() << "WebSocket disconnected from GoRDP API";
}

void GoRDPBridge::onWebSocketTextMessageReceived(const QString &message)
{
    QJsonParseError parseError;
    QJsonDocument doc = QJsonDocument::fromJson(message.toUtf8(), &parseError);
    
    if (parseError.error != QJsonParseError::NoError) {
        qWarning() << "Invalid JSON in WebSocket message:" << parseError.errorString();
        return;
    }
    
    handleWebSocketMessage(doc);
}

void GoRDPBridge::onWebSocketBinaryMessageReceived(const QByteArray &message)
{
    // Handle bitmap data
    QImage image = decodeBitmapData(message);
    if (!image.isNull()) {
        emit bitmapReceived(image);
    }
}

void GoRDPBridge::onProcessFinished(int exitCode, QProcess::ExitStatus exitStatus)
{
    m_apiRunning = false;
    qDebug() << "GoRDP API process finished with exit code:" << exitCode;
    
    if (exitCode != 0) {
        QString errorOutput = m_gordpProcess->readAllStandardError();
        emit apiError(QString("GoRDP API exited with error: %1").arg(errorOutput));
    }
}

void GoRDPBridge::onProcessError(QProcess::ProcessError error)
{
    m_apiRunning = false;
    QString errorString;
    
    switch (error) {
        case QProcess::FailedToStart:
            errorString = "Failed to start GoRDP API";
            break;
        case QProcess::Crashed:
            errorString = "GoRDP API crashed";
            break;
        default:
            errorString = "GoRDP API process error";
            break;
    }
    
    emit apiError(errorString);
}

void GoRDPBridge::sendHttpRequest(const QString &endpoint, const QJsonObject &data)
{
    QUrl url(m_apiUrl + endpoint);
    QNetworkRequest request(url);
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    
    QJsonDocument doc(data);
    m_networkManager->post(request, doc.toJson());
}

void GoRDPBridge::handleHttpResponse(const QJsonDocument &response)
{
    QJsonObject obj = response.object();
    QString type = obj["type"].toString();
    
    if (type == "connection_status") {
        bool connected = obj["connected"].toBool();
        m_isConnected = connected;
        emit connectionStatusChanged(connected);
    } else if (type == "error") {
        emit connectionError(obj["message"].toString());
    } else if (type == "history") {
        emit connectionHistoryLoaded(obj["data"].toArray());
    } else if (type == "favorites") {
        emit favoritesLoaded(obj["data"].toArray());
    } else if (type == "performance") {
        m_lastPerformanceStats = obj["data"].toObject();
        emit performanceStatsReceived(m_lastPerformanceStats);
    }
}

void GoRDPBridge::setupWebSocket()
{
    QString wsUrl = QString("ws://localhost:%1/ws").arg(m_apiPort);
    m_websocket->open(QUrl(wsUrl));
}

void GoRDPBridge::handleWebSocketMessage(const QJsonDocument &message)
{
    QJsonObject obj = message.object();
    QString type = obj["type"].toString();
    
    if (type == "bitmap") {
        // Handle bitmap data (should come as binary message)
        qDebug() << "Bitmap update received";
    } else if (type == "connection_status") {
        bool connected = obj["connected"].toBool();
        m_isConnected = connected;
        emit connectionStatusChanged(connected);
    }
}

QJsonObject GoRDPBridge::createRequest(const QString &action, const QJsonObject &data)
{
    QJsonObject request;
    request["action"] = action;
    if (!data.isEmpty()) {
        request["data"] = data;
    }
    return request;
}

QImage GoRDPBridge::decodeBitmapData(const QByteArray &data)
{
    QImage image;
    if (image.loadFromData(data)) {
        return image;
    }
    
    // Try to decode as raw bitmap data
    // This is a simplified implementation - actual RDP bitmap decoding
    // would be more complex and handled by the GoRDP core
    qWarning() << "Failed to decode bitmap data";
    return QImage();
}

QString GoRDPBridge::getGoRDPExecutablePath()
{
    // First check if gordp-api is in PATH
    QProcess process;
    process.start("which", QStringList() << "gordp-api");
    if (process.waitForFinished() && process.exitCode() == 0) {
        return process.readAllStandardOutput().trimmed();
    }
    
    // Check common locations
    QStringList possiblePaths = {
        "./gordp-api",
        "../gordp-api",
        "bin/gordp-api",
        QStandardPaths::findExecutable("gordp-api")
    };
    
    for (const QString &path : possiblePaths) {
        if (QFile::exists(path)) {
            return path;
        }
    }
    
    return QString();
} 