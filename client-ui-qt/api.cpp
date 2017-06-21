#include "api.h"

QString BASE_URL = "http://localhost:61337";

API::API() :
    qnam(new QNetworkAccessManager())
{

}

QNetworkReply* API::GetStatus() {
    return qnam->get(QNetworkRequest(QUrl(QString("%1%2").arg(BASE_URL).arg("/api/status"))));
}

QNetworkReply* API::GetShards() {
    return qnam->get(QNetworkRequest(QUrl(QString("%1%2").arg(BASE_URL).arg("/api/shards"))));
}

QNetworkReply* API::ShardRelay(QString shardAddr, QString transport) {
    return qnam->get(QNetworkRequest(QUrl(QString("%1%2?shardAddr=%3&transport=%4").arg(BASE_URL).arg("/api/shards/relay").arg(shardAddr).arg(transport))));
}

QNetworkReply* API::GetConnectionStatus() {
    return qnam->get(QNetworkRequest(QUrl(QString("%1%2").arg(BASE_URL).arg("/api/connectionStatus"))));
}
