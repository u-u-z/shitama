#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
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
    QSettings* settings;
    API* api;
    int statusWatcher;
    QProcess* clientProcess;
    QProcess* childProcess;

    virtual void timerEvent(QTimerEvent* event);

    void startClient();
    void updateStatus();
    void updateShards();
    void shardRelay();
    void copyAddress();
    void launchChild();
    void stopClient();
};

#endif // MAINWINDOW_H
