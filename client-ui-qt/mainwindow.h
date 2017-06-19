#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QSystemTrayIcon>
#include <QProcess>
#include <QTimer>
#include <QSettings>

#include <api.h>

namespace Ui {
class MainWindow;
}

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget* parent = 0);
    ~MainWindow();

private:
    Ui::MainWindow* ui;
    QSystemTrayIcon* tray;
    QSettings* settings;
    API* api;
    int statusWatcher;
    QProcess* clientProcess;
    QProcess* childProcess;

    bool connected;

    virtual void timerEvent(QTimerEvent* event);

    void setConnected(bool connected, bool forceUpdate = false);

    void startClient();
    void updateStatus();
    void updateShards();
    void shardRelay();
    void copyAddress();
    void launch();
    void stopClient();
};

#endif // MAINWINDOW_H
