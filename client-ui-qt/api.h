#ifndef API_H
#define API_H

#include <QString>
#include <QtNetwork/QNetworkAccessManager>
#include <QJsonObject>
#include <QJsonArray>

class API
{
public:
    API();
    QJsonObject GetStatus();
    QJsonArray GetShards();
    QJsonObject ShardRelay(QString shardAddr, QString transport);
private:
    QNetworkAccessManager* qnam;
};

#endif // API_H
