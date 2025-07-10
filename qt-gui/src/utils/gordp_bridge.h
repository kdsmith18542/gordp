#ifndef GORDP_BRIDGE_H
#define GORDP_BRIDGE_H

#include <QObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QWebSocket>
#include <QJsonObject>
#include <QJsonDocument>
#include <QJsonArray>
#include <QTimer>
#include <QProcess>
#include <QImage>
#include <QByteArray>
#include <QUrl>
#include <QString>

class GoRDPBridge : public QObject
{
    Q_OBJECT

public:
    explicit GoRDPBridge(QObject *parent = nullptr);
    ~GoRDPBridge();

    // Connection management
    bool checkGoRDPAvailability();
    void startGoRDPAPI();
    void stopGoRDPAPI();
    
    // RDP connection methods
    void connectToServer(const QString &server, int port, 
                        const QString &username, const QString &password,
                        const QJsonObject &options = QJsonObject());
    void disconnectFromServer();
    
    // Input methods
    void sendMouseEvent(int x, int y, int button, bool pressed);
    void sendKeyEvent(int key, bool pressed);
    void sendWheelEvent(int delta);
    
    // Settings and data
    void saveConnectionHistory(const QJsonObject &connection);
    void loadConnectionHistory();
    void saveFavorites(const QJsonArray &favorites);
    void loadFavorites();
    
    // Performance monitoring
    void getPerformanceStats();
    void getConnectionStatus();

signals:
    // Connection events
    void connectionStatusChanged(bool connected);
    void connectionError(const QString &error);
    void bitmapReceived(const QImage &image);
    void errorOccurred(const QString &error);
    
    // Data events
    void connectionHistoryLoaded(const QJsonArray &history);
    void favoritesLoaded(const QJsonArray &favorites);
    void performanceStatsReceived(const QJsonObject &stats);
    
    // API events
    void apiStarted();
    void apiStopped();
    void apiError(const QString &error);

private slots:
    void onNetworkReplyFinished(QNetworkReply *reply);
    void onWebSocketConnected();
    void onWebSocketDisconnected();
    void onWebSocketTextMessageReceived(const QString &message);
    void onWebSocketBinaryMessageReceived(const QByteArray &message);
    void onProcessFinished(int exitCode, QProcess::ExitStatus exitStatus);
    void onProcessError(QProcess::ProcessError error);

private:
    // HTTP API methods
    void sendHttpRequest(const QString &endpoint, const QJsonObject &data = QJsonObject());
    void handleHttpResponse(const QJsonDocument &response);
    
    // WebSocket methods
    void setupWebSocket();
    void handleWebSocketMessage(const QJsonDocument &message);
    
    // Utility methods
    QJsonObject createRequest(const QString &action, const QJsonObject &data = QJsonObject());
    QImage decodeBitmapData(const QByteArray &data);
    QString getGoRDPExecutablePath();
    
    // Network components
    QNetworkAccessManager *m_networkManager;
    QWebSocket *m_websocket;
    QProcess *m_gordpProcess;
    
    // Configuration
    QString m_apiUrl;
    int m_apiPort;
    bool m_isConnected;
    bool m_apiRunning;
    
    // Connection info
    QString m_currentServer;
    int m_currentPort;
    QString m_currentUsername;
    
    // Performance tracking
    QTimer *m_performanceTimer;
    QJsonObject m_lastPerformanceStats;
};

#endif // GORDP_BRIDGE_H 