#ifndef API_H
#define API_H

#include <QString>
#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkReply>

class API
{
public:
    API();
    QNetworkReply* GetStatus();
    QNetworkReply* GetShards();
    QNetworkReply* ShardRelay(QString shardAddr, QString transport);
    QNetworkReply* GetConnectionStatus();
private:
    QNetworkAccessManager* qnam;
};

#endif // API_H
