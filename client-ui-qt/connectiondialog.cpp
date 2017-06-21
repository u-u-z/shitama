#include "connectiondialog.h"
#include "ui_connectiondialog.h"

#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QSplineSeries>
#include <QValueAxis>

ConnectionDialog::ConnectionDialog(QWidget *parent) :
    QDialog(parent),
    ui(new Ui::ConnectionDialog),
    chartView(new QtCharts::QChartView()),
    chart(new QtCharts::QChart()),
    chartIndex(0),
    api(new API())
{

    this->ui->setupUi(this);

    this->chartView->setChart(this->chart);

    this->ui->formLayout->addRow(tr("可视化延迟"), this->chartView);

    this->statusWatcher = this->startTimer(1000);

}

void ConnectionDialog::timerEvent(QTimerEvent* event) {

    this->updateConnectionStatus();

}

void ConnectionDialog::updateConnectionStatus() {

    QNetworkReply* reply = api->GetConnectionStatus();
    connect(reply, &QNetworkReply::finished, [=]() {

        this->chartIndex++;

        // New QMap for filtering inactive peers.
        // FIXME: Better solution? This one might have performance issues.
        QMap<QString, QList<qreal>> chartData;

        auto status = QJsonDocument::fromJson(reply->readAll()).object();
        auto linkEstablished = status["linkEstablished"].toBool();
        auto linkAddr = status["linkAddr"].toString();
        auto linkDelay = status["linkDelay"].toInt();
        auto linkDelayDelta = status["linkDelayDelta"].toInt();

        auto linkData = this->chartData["link"];
        linkData.push_front(double(linkDelay) / 1000000);
        chartData["link"] = linkData;

        this->ui->linkStatus->setText(linkEstablished ? tr("已建立") : tr("未建立"));
        this->ui->linkAddr->setText(linkAddr);
        this->ui->linkDelay->setText(tr("%1ms ± %2ms").arg(double(linkDelay) / 1000000).arg(double(linkDelayDelta) / 1000000));

        auto peers = status["peers"].toArray();

        this->ui->peers->clear();

        for(int i = 0; i < peers.size(); i++) {

            auto peer = peers.at(i).toObject();

            auto remoteAddr = peer["remoteAddr"].toString();
            auto localAddr = peer["localAddr"].toString();
            auto delay = peer["delay"].toInt();
            auto profile = peer["profile"].toString();
            auto active = peer["active"].toInt();

            if(delay != 0) {

                auto data = this->chartData[remoteAddr];
                data.push_front(double(delay) / 1000000);
                chartData[remoteAddr] = data;

                this->ui->peers->addItem(tr("%1 - %2ms - %3").arg(remoteAddr).arg(double(delay) / 1000000).arg(profile));

            }

        }

        this->chartData = chartData;

        this->updateDelayChart();

    });

}

void ConnectionDialog::updateDelayChart() {

    this->chart->removeAllSeries();

    for(auto it = this->chartData.begin(); it != this->chartData.end(); it++) {

        // These series will be released by QtCharts::QChart::removeAllSeries(), so it can't be reused.
        auto series = new QtCharts::QLineSeries();
        series->setName(it.key());

        for(int i = 0; i < 32; i++) {
            series->append(i, it.value().value(i));
        }

        this->chart->addSeries(series);

    }

    // Also there is a need for reseting the ranges.
    this->chart->createDefaultAxes();
    this->chart->axisX()->setRange(0, 31);
    this->chart->axisX()->hide();
    this->chart->axisY()->setRange(0, 100);

    /*
    auto xAxis = new QtCharts::QValueAxis();
    xAxis->setRange(0, 32);
    this->chart->setAxisX(xAxis);

    auto yAxis = new QtCharts::QValueAxis();
    yAxis->setRange(0, 100);
    this->chart->setAxisY(yAxis);
    */

}

ConnectionDialog::~ConnectionDialog()
{
    delete this->ui;
}
