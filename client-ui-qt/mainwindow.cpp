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
    settings(new QSettings()),
    api(new API()),
    clientProcess(nullptr),
    childProcess(nullptr),
    connected(false)
{

    ui->setupUi(this);

    this->setWindowFlags(this->windowFlags() & ~Qt::WindowMinMaxButtonsHint);

    connect(ui->actionClearConfig, &QAction::triggered, [=]() {
        this->settings->clear();
        QMessageBox::information(nullptr, this->windowTitle(), tr("应用配置已清空。"));
    });
    connect(ui->actionExit, &QAction::triggered, this, &MainWindow::close);
    connect(ui->actionCheckUpdates, &QAction::triggered, [=]() {
        QDesktopServices::openUrl(QUrl("https://github.com/evshiron/shitama/releases"));
    });

    connect(ui->updateShards, &QPushButton::clicked, this, &MainWindow::updateShards);
    connect(ui->shardRelay, &QPushButton::clicked, this, &MainWindow::shardRelay);
    connect(ui->copyAddress, &QPushButton::clicked, this, &MainWindow::copyAddress);
    connect(ui->launch, &QPushButton::clicked, this, &MainWindow::launch);

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
            ui->shards->setEnabled(true);
            ui->updateShards->setEnabled(true);
            ui->address->setEnabled(true);
            ui->shardRelay->setEnabled(true);
            ui->copyAddress->setEnabled(true);

            this->updateShards();
        }
        else {
            ui->shards->setDisabled(true);
            ui->address->setDisabled(true);
            ui->updateShards->setDisabled(true);
            ui->shardRelay->setDisabled(true);
            ui->copyAddress->setDisabled(true);
        }
    }

    this->connected = connected;

    if(this->connected) {
        ui->statusBar->showMessage(tr("已连接。"));
    }
    else {
        ui->statusBar->showMessage(tr("已断开。"));
    }

}

void MainWindow::startClient() {

    #ifdef Q_OS_WIN32
    QString path = ".\\client.exe";
    #else
    QString path = "./client";
    #endif

    QFileInfo file(path);

    if(!file.exists() || !file.isFile()) {

        QMessageBox::warning(nullptr, this->windowTitle(), tr("文件「%1」缺失。").arg(path));

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

    ui->shards->setDisabled(true);
    ui->updateShards->setDisabled(true);

    QNetworkReply* reply = api->GetShards();
    connect(reply, &QNetworkReply::finished, [=]() {

        auto shards = QJsonDocument::fromJson(reply->readAll()).array();

        ui->shards->clear();

        for(int i = 0; i < shards.size(); i++) {

            auto shard = shards.at(i).toObject();

            ui->shards->addItem(QString("%1 - %2ms").arg(shard["ip"].toString()).arg(shard["rtt"].toDouble(), 0, 'f', 2), QVariant(shard["addr"].toString()));

        }

        ui->shards->setEnabled(true);
        ui->updateShards->setEnabled(true);

    });

}

void MainWindow::shardRelay() {

    ui->address->setDisabled(true);
    ui->shardRelay->setDisabled(true);

    auto shardAddr = ui->shards->currentData().toString();
    auto transport = ui->transports->currentText().toLower();

    QNetworkReply* reply = api->ShardRelay(shardAddr, transport);
    connect(reply, &QNetworkReply::finished, [=]() {

        auto relayInfo = QJsonDocument::fromJson(reply->readAll()).object();

        ui->address->setText(relayInfo["guestAddr"].toString());

        ui->address->setEnabled(true);
        ui->shardRelay->setEnabled(true);

    });

}

void MainWindow::copyAddress() {

    ui->copyAddress->setDisabled(true);

    QClipboard* clipboard = QApplication::clipboard();

    QString address = ui->address->text().trimmed();

    if(address == "") {
        QMessageBox::information(nullptr, this->windowTitle(), tr("请先点击「中转」按钮获取地址。"));
    }
    else {

        clipboard->setText(address);
        QMessageBox::information(nullptr, this->windowTitle(), tr("「%1」已经复制到剪贴板。").arg(address));

    }

    ui->copyAddress->setEnabled(true);

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
    delete this->settings;
    delete this->api;

    delete this->childProcess;

}
