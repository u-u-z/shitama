#include "mainwindow.h"
#include "ui_mainwindow.h"

#include <QMessageBox>
#include <QFileDialog>
#include <QFileInfo>
#include <QJsonDocument>
#include <QClipboard>

MainWindow::MainWindow(QWidget *parent) :
    QMainWindow(parent),
    ui(new Ui::MainWindow),
    settings(new QSettings()),
    api(new API()),
    clientProcess(nullptr),
    childProcess(nullptr)
{

    ui->setupUi(this);

    this->setWindowFlags(this->windowFlags() & ~Qt::WindowMinMaxButtonsHint);

    connect(ui->updateShards, &QPushButton::clicked, this, &MainWindow::updateShards);
    connect(ui->shardRelay, &QPushButton::clicked, this, &MainWindow::shardRelay);
    connect(ui->copyAddress, &QPushButton::clicked, this, &MainWindow::copyAddress);
    connect(ui->launch, &QPushButton::clicked, this, &MainWindow::launchChild);

    this->startClient();

    ui->statusBar->showMessage(tr("初始化。"));

    this->statusWatcher = this->startTimer(1000);

    this->updateShards();

}

void MainWindow::timerEvent(QTimerEvent* event) {

    this->updateStatus();

}

void MainWindow::startClient() {

    #ifdef Q_OS_WIN32
    QString path = ".\\client.exe";
    #else
    QString path = "./client";
    #endif

    QFileInfo file(path);

    if(!file.exists() || !file.isFile()) {

        QMessageBox msgBox;
        msgBox.setWindowTitle(tr("Shitama"));
        msgBox.setText(tr("文件「%1」缺失。").arg(path));
        msgBox.exec();

    }
    else {

        this->clientProcess = new QProcess(this);
        this->clientProcess->start(path);
        this->clientProcess->waitForStarted();

    }

}

void MainWindow::updateStatus() {

    auto status = api->GetStatus();

    auto connected = status["connected"].toBool();

    if(connected) {
        ui->statusBar->showMessage(tr("已连接。"));
    }
    else {
        ui->statusBar->showMessage(tr("已断开。"));
    }

}

void MainWindow::updateShards() {

    ui->updateShards->setEnabled(false);

    auto shards = api->GetShards();

    ui->shards->clear();

    for(int i = 0; i < shards.size(); i++) {

        auto shard = shards.at(i).toObject();

        ui->shards->addItem(QString("%1 - %2ms").arg(shard["ip"].toString()).arg(shard["rtt"].toDouble(), 0, 'f', 2), QVariant(shard["addr"].toString()));

    }

    ui->updateShards->setEnabled(true);

}

void MainWindow::shardRelay() {

    ui->shardRelay->setEnabled(false);

    auto shardAddr = ui->shards->currentData().toString();
    auto transport = ui->transports->currentText().toLower();

    auto relayInfo = api->ShardRelay(shardAddr, transport);

    ui->address->setText(relayInfo["guestAddr"].toString());

    ui->shardRelay->setEnabled(true);

}

void MainWindow::copyAddress() {

    ui->copyAddress->setEnabled(false);

    QClipboard* clipboard = QApplication::clipboard();

    QString address = ui->address->text().trimmed();

    if(address == "") {

        QMessageBox msgBox;
        msgBox.setWindowTitle(tr("Shitama"));
        msgBox.setText(tr("请先点击「中转」按钮获取地址。"));
        msgBox.exec();

    }
    else {

        clipboard->setText(address);

        QMessageBox msgBox;
        msgBox.setWindowTitle(tr("Shitama"));
        msgBox.setText(tr("「%1」已经复制到剪贴板。").arg(address));
        msgBox.exec();

    }

    ui->copyAddress->setEnabled(true);

}

void MainWindow::launchChild() {

    auto path = this->settings->value("launch/path", QVariant("")).toString();

    if(path == "") {

        path = QFileDialog::getOpenFileName(this, tr("Shitama"));
        this->settings->setValue("launch/path", QVariant(path));

    }

    this->childProcess = new QProcess(this);
    this->childProcess->start(path);
    this->childProcess->waitForStarted();

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
