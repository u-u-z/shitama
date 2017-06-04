#include "api.h"

#include <QEventLoop>
#include <QtNetwork/QNetworkReply>
#include <QJsonDocument>

QString BASE_URL = "http://localhost:61337";

API::API() :
    qnam(new QNetworkAccessManager())
{

}

QJsonObject API::GetStatus() {

    auto reply = qnam->get(QNetworkRequest(QUrl(QString("%1%2").arg(BASE_URL).arg("/api/status"))));

    QEventLoop loop;
    reply->connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();

    auto data = reply->readAll();
    //qInfo() << data;

    auto json = QJsonDocument::fromJson(data);
    return json.object();

}

QJsonArray API::GetShards() {

    auto reply = qnam->get(QNetworkRequest(QUrl(QString("%1%2").arg(BASE_URL).arg("/api/shards"))));

    QEventLoop loop;
    reply->connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();

    auto data = reply->readAll();
    qInfo() << data;

    auto json = QJsonDocument::fromJson(data);
    return json.array();

}

QJsonObject API::ShardRelay(QString shardAddr, QString transport) {

    auto reply = qnam->get(QNetworkRequest(QUrl(QString("%1%2?shardAddr=%3&transport=%4").arg(BASE_URL).arg("/api/shards/relay").arg(shardAddr).arg(transport))));

    QEventLoop loop;
    reply->connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();

    auto data = reply->readAll();
    qInfo() << data;

    auto json = QJsonDocument::fromJson(data);
    return json.object();

}
