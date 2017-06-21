#include "mainwindow.h"
#include "ui_mainwindow.h"

#include <QMessageBox>
#include <QFileDialog>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QClipboard>
#include <QDesktopServices>

MainWindow::MainWindow(QWidget *parent) :
    QMainWindow(parent),
    ui(new Ui::MainWindow),
    connectionDialog(new ConnectionDialog()),
    tray(new QSystemTrayIcon()),
    settings(new QSettings()),
    api(new API()),
    clientProcess(nullptr),
    childProcess(nullptr),
    connected(false)
{

    this->ui->setupUi(this);

    this->tray->setToolTip(this->windowTitle());
    this->tray->setIcon(QIcon(":/shitama.ico"));
    this->tray->show();

    this->setWindowFlags(this->windowFlags() & ~Qt::WindowMinMaxButtonsHint);

    connect(this->ui->actionClearConfig, &QAction::triggered, [=]() {
        this->settings->clear();
        this->tray->showMessage(this->windowTitle(), tr("应用配置已清空。"), QSystemTrayIcon::Information, 1000);
    });
    connect(this->ui->actionExit, &QAction::triggered, this, &MainWindow::close);
    connect(this->ui->actionConnectionView, &QAction::triggered, [=]() {
        this->connectionDialog->show();
    });
    connect(this->ui->actionCheckUpdates, &QAction::triggered, [=]() {
        QDesktopServices::openUrl(QUrl("https://github.com/evshiron/shitama/releases"));
    });
    connect(this->ui->actionAbout, &QAction::triggered, [=]() {
        QMessageBox::about(nullptr, this->windowTitle(), tr("Shitama r%1\r\n"
                                                            "Commit: %2").arg(SHITAMA_BUILD_ID).arg(SHITAMA_COMMIT));
    });

    connect(this->ui->updateShards, &QPushButton::clicked, this, &MainWindow::updateShards);
    connect(this->ui->shardRelay, &QPushButton::clicked, this, &MainWindow::shardRelay);
    connect(this->ui->copyAddress, &QPushButton::clicked, this, &MainWindow::copyAddress);
    connect(this->ui->launch, &QPushButton::clicked, this, &MainWindow::launch);

    this->startClient();

    this->statusWatcher = this->startTimer(1000);

    this->setConnected(false, true);

}

void MainWindow::timerEvent(QTimerEvent* event) {

    this->updateStatus();

}

void MainWindow::setConnected(bool connected, bool forceUpdate) {

    if(this->connected != connected || forceUpdate) {
        if(connected) {

            this->ui->shards->setEnabled(true);
            this->ui->updateShards->setEnabled(true);
            this->ui->address->setEnabled(true);
            this->ui->shardRelay->setEnabled(true);
            this->ui->copyAddress->setEnabled(true);

            if(!forceUpdate) {
                this->tray->showMessage(this->windowTitle(), tr("服务器已连接。"), QSystemTrayIcon::Information, 1000);
            }

            this->updateShards();
        }
        else {

            this->ui->shards->setDisabled(true);
            this->ui->address->setDisabled(true);
            this->ui->updateShards->setDisabled(true);
            this->ui->shardRelay->setDisabled(true);
            this->ui->copyAddress->setDisabled(true);

            if(!forceUpdate) {
                this->tray->showMessage(this->windowTitle(), tr("服务器已断开。"), QSystemTrayIcon::Information, 1000);
            }

        }
    }

    this->connected = connected;

}

void MainWindow::startClient() {

    #ifdef Q_OS_WIN32
    QString path = ".\\client.exe";
    #else
    QString path = "./client";
    #endif

    QFileInfo file(path);

    if(!file.exists() || !file.isFile()) {
        this->tray->showMessage(this->windowTitle(), tr("文件「%1」缺失。").arg(path), QSystemTrayIcon::Warning, 1000);
    }
    else {

        this->clientProcess = new QProcess(this);
        this->clientProcess->start(path);
        this->clientProcess->waitForStarted();

    }

}

void MainWindow::updateStatus() {

    QNetworkReply* reply = api->GetStatus();
    connect(reply, &QNetworkReply::finished, [=]() {

        auto status = QJsonDocument::fromJson(reply->readAll()).object();
        auto connected = status["connected"].toBool();

        this->setConnected(connected);

    });

}

void MainWindow::updateShards() {

    this->ui->shards->setDisabled(true);
    this->ui->updateShards->setDisabled(true);

    QNetworkReply* reply = api->GetShards();
    connect(reply, &QNetworkReply::finished, [=]() {

        auto shards = QJsonDocument::fromJson(reply->readAll()).array();

        this->ui->shards->clear();

        for(int i = 0; i < shards.size(); i++) {

            auto shard = shards.at(i).toObject();

            this->ui->shards->addItem(QString("%1 - %2ms").arg(shard["ip"].toString()).arg(shard["rtt"].toDouble(), 0, 'f', 2), QVariant(shard["addr"].toString()));

        }

        this->ui->shards->setEnabled(true);
        this->ui->updateShards->setEnabled(true);

    });

}

void MainWindow::shardRelay() {

    this->ui->address->setDisabled(true);
    this->ui->shardRelay->setDisabled(true);

    auto shardAddr = this->ui->shards->currentData().toString();
    auto transport = this->ui->transports->currentText().toLower();

    QNetworkReply* reply = api->ShardRelay(shardAddr, transport);
    connect(reply, &QNetworkReply::finished, [=]() {

        auto relayInfo = QJsonDocument::fromJson(reply->readAll()).object();

        this->ui->address->setText(relayInfo["guestAddr"].toString());

        this->ui->address->setEnabled(true);
        this->ui->shardRelay->setEnabled(true);

    });

}

void MainWindow::copyAddress() {

    this->ui->copyAddress->setDisabled(true);

    QClipboard* clipboard = QApplication::clipboard();

    QString address = this->ui->address->text().trimmed();

    if(address == "") {
        this->tray->showMessage(this->windowTitle(), tr("请先点击「中转」按钮获取地址。"), QSystemTrayIcon::Warning, 1000);
    }
    else {

        clipboard->setText(address);
        this->tray->showMessage(this->windowTitle(), tr("「%1」已经复制到剪贴板。").arg(address), QSystemTrayIcon::Information, 1000);

    }

    this->ui->copyAddress->setEnabled(true);

}

void MainWindow::launch() {

    auto path = this->settings->value("launch/path", QVariant("")).toString();

    if(path == "") {

        path = QFileDialog::getOpenFileName(nullptr, this->windowTitle());
        this->settings->setValue("launch/path", QVariant(path));

    }

    this->ui->launch->setDisabled(true);

    this->childProcess = new QProcess(this);
    this->childProcess->start(path);
    this->childProcess->waitForStarted();

    connect(this->childProcess, static_cast<void (QProcess::*) (int, QProcess::ExitStatus)>(&QProcess::finished), [=](int exitCode, QProcess::ExitStatus exitStatus) {
        this->ui->launch->setEnabled(true);
    });

}

void MainWindow::stopClient() {

    if(this->clientProcess != nullptr) {

        this->clientProcess->kill();
        this->clientProcess = nullptr;

    }

}

MainWindow::~MainWindow()
{

    this->killTimer(this->statusWatcher);
    this->statusWatcher = 0;

    this->stopClient();

    delete this->ui;
    delete this->tray;
    delete this->settings;
    delete this->api;

    delete this->childProcess;

}
