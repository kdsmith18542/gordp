#include <QApplication>
#include <QMainWindow>
#include <QStyleFactory>
#include <QDir>
#include <QStandardPaths>
#include <QMessageBox>
#include <QProcess>
#include <QTimer>

#include "mainwindow/mainwindow.h"
#include "utils/gordp_bridge.h"

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    
    // Set application metadata
    app.setApplicationName("GoRDP GUI");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("GoRDP Project");
    app.setOrganizationDomain("gordp.org");
    
    // Set application style
    app.setStyle(QStyleFactory::create("Fusion"));
    
    // Create main window
    MainWindow window;
    window.show();
    
    // Check if GoRDP core is available
    QTimer::singleShot(100, [&window]() {
        GoRDPBridge bridge;
        if (!bridge.checkGoRDPAvailability()) {
            QMessageBox::warning(&window, "GoRDP Core Not Found", 
                "GoRDP core executable not found. Please ensure gordp-api is available in PATH.\n\n"
                "The GUI will start but RDP connections will not work.");
        }
    });
    
    return app.exec();
} 